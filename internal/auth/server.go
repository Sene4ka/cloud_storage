package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/repositories"
	"github.com/Sene4ka/cloud_storage/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	api.UnimplementedAuthServiceServer
	userRepo    *repositories.UserRepository
	jwtManager  *utils.JWTManager
	redisClient *redis.Client
	config      *configs.Config
}

func NewServer(userRepo *repositories.UserRepository, redisClient *redis.Client, config *configs.Config) *Server {
	return &Server{
		userRepo:    userRepo,
		jwtManager:  utils.NewJWTManager(config.JWT.Secret, config.JWT.AccessTokenTTL, config.JWT.RefreshTokenTTL),
		redisClient: redisClient,
		config:      config,
	}
}

func (s *Server) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user with this email already exists")
	}

	user, err := models.NewUser(req.Email, req.Password, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.redisClient.Set(ctx, "refresh:"+user.ID, refreshToken, s.config.JWT.RefreshTokenTTL).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &api.RegisterResponse{
		UserId:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		AccessToken:      accessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.redisClient.Set(ctx, "refresh:"+user.ID, refreshToken, s.config.JWT.RefreshTokenTTL).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &api.LoginResponse{
		UserId:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		AccessToken:      accessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *api.RefreshRequest) (*api.RefreshResponse, error) {
	blacklisted, err := s.redisClient.Exists(ctx, "blacklist:"+req.RefreshToken).Result()
	if err != nil {
		return nil, fmt.Errorf("redis blacklist check failed: %w", err)
	}
	if blacklisted > 0 {
		return nil, fmt.Errorf("refresh token blacklisted")
	}

	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	storedRefresh, err := s.redisClient.Get(ctx, "refresh:"+claims.UserID).Result()
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}
	if storedRefresh != req.RefreshToken {
		return nil, fmt.Errorf("refresh token mismatch")
	}

	newAccessToken, newRefreshToken, err := s.jwtManager.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	err = s.redisClient.Set(ctx, "refresh:"+claims.UserID, newRefreshToken, s.config.JWT.RefreshTokenTTL).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return &api.RefreshResponse{
		AccessToken:      newAccessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     newRefreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *api.ValidateTokenRequest) (*api.ValidateTokenResponse, error) {
	claims, err := s.jwtManager.ValidateAccessToken(req.Token)
	if err != nil {
		return &api.ValidateTokenResponse{Valid: false}, nil
	}

	return &api.ValidateTokenResponse{
		Valid:     true,
		UserId:    claims.UserID,
		Email:     claims.Email,
		ExpiresIn: int64(time.Until(claims.ExpiresAt.Time).Seconds()),
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *api.LogoutRequest) (*api.LogoutResponse, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return &api.LogoutResponse{Success: false}, nil
	}

	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL > 0 {
		err = s.redisClient.Set(ctx, "blacklist:"+req.RefreshToken, "1", remainingTTL).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to blacklist token: %w", err)
		}
	}

	err = s.redisClient.Del(ctx, "refresh:"+claims.UserID).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return &api.LogoutResponse{Success: true}, nil
}
