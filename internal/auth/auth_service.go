package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/utils"
	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type TokenCache interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (int64, error)
}

type TokenManager interface {
	GenerateTokenPair(userID, email string) (string, string, error)
	ValidateAccessToken(token string) (*utils.TokenClaims, error)
	ValidateRefreshToken(token string) (*utils.TokenClaims, error)
}

type authService struct {
	userRepo   UserRepository
	tokenCache TokenCache
	tokenMgr   TokenManager
	config     *configs.Config
}

func NewAuthService(userRepo UserRepository, tokenCache TokenCache, tokenMgr TokenManager, config *configs.Config) *authService {
	return &authService{
		userRepo:   userRepo,
		tokenCache: tokenCache,
		tokenMgr:   tokenMgr,
		config:     config,
	}
}

func NewAuthServiceRedis(userRepo UserRepository, redisClient *redis.Client, config *configs.Config) *authService {
	tokenCache := NewRedisAdapter(redisClient)
	tokenMgr := utils.NewJWTManager(config.JWT.Secret, config.JWT.AccessTokenTTL, config.JWT.RefreshTokenTTL)
	return NewAuthService(userRepo, tokenCache, tokenMgr, config)
}

func (s *authService) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with this email already exists")
	}

	user, err := models.NewUser(input.Email, input.Password, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	accessToken, refreshToken, err := s.tokenMgr.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.tokenCache.Set(ctx, "refresh:"+user.ID, refreshToken, s.config.JWT.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &RegisterOutput{
		UserID:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		AccessToken:      accessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *authService) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.CheckPassword(input.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, refreshToken, err := s.tokenMgr.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.tokenCache.Set(ctx, "refresh:"+user.ID, refreshToken, s.config.JWT.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &LoginOutput{
		UserID:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		AccessToken:      accessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *authService) Refresh(ctx context.Context, input *RefreshInput) (*RefreshOutput, error) {
	blacklisted, err := s.tokenCache.Exists(ctx, "blacklist:"+input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("redis blacklist check failed: %w", err)
	}
	if blacklisted > 0 {
		return nil, fmt.Errorf("refresh token blacklisted")
	}

	claims, err := s.tokenMgr.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	storedRefresh, err := s.tokenCache.Get(ctx, "refresh:"+claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}
	if storedRefresh != input.RefreshToken {
		return nil, fmt.Errorf("refresh token mismatch")
	}

	newAccessToken, newRefreshToken, err := s.tokenMgr.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	err = s.tokenCache.Set(ctx, "refresh:"+claims.UserID, newRefreshToken, s.config.JWT.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return &RefreshOutput{
		AccessToken:      newAccessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     newRefreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, input *ValidateTokenInput) (*ValidateTokenOutput, error) {
	claims, err := s.tokenMgr.ValidateAccessToken(input.Token)
	if err != nil {
		return &ValidateTokenOutput{Valid: false}, nil
	}

	return &ValidateTokenOutput{
		Valid:     true,
		UserID:    claims.UserID,
		Email:     claims.Email,
		ExpiresIn: int64(time.Until(claims.ExpiresAt.Time).Seconds()),
	}, nil
}

func (s *authService) Logout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error) {
	claims, err := s.tokenMgr.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return &LogoutOutput{Success: false}, nil
	}

	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL > 0 {
		err = s.tokenCache.Set(ctx, "blacklist:"+input.RefreshToken, "1", remainingTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to blacklist token: %w", err)
		}
	}

	err = s.tokenCache.Del(ctx, "refresh:"+claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return &LogoutOutput{Success: true}, nil
}
