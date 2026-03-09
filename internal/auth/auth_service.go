package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/Sene4ka/cloud_storage/internal/api"
	"github.com/Sene4ka/cloud_storage/internal/models"
	"github.com/Sene4ka/cloud_storage/internal/utils"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Update(ctx context.Context, user *models.User) error
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
	GenerateTempToken(userID, email string) (string, error)
	ValidateTempToken(tokenString string) (*utils.TokenClaims, error)
}

type MailService interface {
	Send2FACode(ctx context.Context, input *api.Send2FACodeRequest, opts ...grpc.CallOption) (*api.Send2FACodeResponse, error)
}

type authService struct {
	userRepo   UserRepository
	tokenCache TokenCache
	tokenMgr   TokenManager
	mailSvc    MailService
	config     *configs.Config
}

func NewAuthService(userRepo UserRepository, tokenCache TokenCache, tokenMgr TokenManager, mailSvc MailService, config *configs.Config) *authService {
	return &authService{
		userRepo:   userRepo,
		tokenCache: tokenCache,
		tokenMgr:   tokenMgr,
		mailSvc:    mailSvc,
		config:     config,
	}
}

func NewAuthServiceRedis(userRepo UserRepository, redisClient *redis.Client, mailSvc MailService, config *configs.Config) *authService {
	tokenCache := NewRedisAdapter(redisClient)
	tokenMgr := utils.NewJWTManager(config.JWT.Secret, config.JWT.AccessTokenTTL, config.JWT.RefreshTokenTTL)

	return NewAuthService(userRepo, tokenCache, tokenMgr, mailSvc, config)
}

func generate2FACode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
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

	code, err := generate2FACode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	_, err = s.mailSvc.Send2FACode(ctx, &api.Send2FACodeRequest{
		EmailAddress: user.Email,
		Code:         code,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	err = s.tokenCache.Set(ctx, "verify:"+user.ID, code, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to store verification code: %w", err)
	}

	return &RegisterOutput{
		UserID:  user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Message: "Verification code sent to email",
	}, nil
}

func (s *authService) RegisterComplete(ctx context.Context, input *RegisterCompleteInput) (*RegisterCompleteOutput, error) {
	storedCode, err := s.tokenCache.Get(ctx, "verify:"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("verification code not found or expired")
	}

	if storedCode != input.Code {
		return nil, fmt.Errorf("invalid verification code")
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.IsVerified = true
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	_ = s.tokenCache.Del(ctx, "verify:"+input.UserID)

	accessToken, refreshToken, err := s.tokenMgr.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.tokenCache.Set(ctx, "refresh:"+user.ID, refreshToken, s.config.JWT.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &RegisterCompleteOutput{
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

	if user.Is2FAEnabled || !user.IsVerified {
		code, err := generate2FACode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate 2FA code: %w", err)
		}

		_, err = s.mailSvc.Send2FACode(ctx, &api.Send2FACodeRequest{
			EmailAddress: user.Email,
			Code:         code,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send 2FA code: %w", err)
		}

		err = s.tokenCache.Set(ctx, "2fa:"+user.ID, code, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to store 2FA code: %w", err)
		}

		tempToken, err := s.tokenMgr.GenerateTempToken(user.ID, user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp token: %w", err)
		}

		return &LoginOutput{
			UserID:      user.ID,
			Email:       user.Email,
			Name:        user.Name,
			TempToken:   tempToken,
			Requires2FA: true,
			Message:     "2FA code sent to email",
		}, nil
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
		Requires2FA:      false,
	}, nil
}

func (s *authService) LoginComplete(ctx context.Context, input *LoginCompleteInput) (*LoginCompleteOutput, error) {
	claims, err := s.tokenMgr.ValidateTempToken(input.TempToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired temp token: %w", err)
	}

	storedCode, err := s.tokenCache.Get(ctx, "2fa:"+claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("2FA code not found or expired")
	}

	if storedCode != input.Code {
		return nil, fmt.Errorf("invalid 2FA code")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsVerified {
		user.IsVerified = true
		user.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user verification: %w", err)
		}
	}

	accessToken, refreshToken, err := s.tokenMgr.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	err = s.tokenCache.Set(ctx, "refresh:"+user.ID, refreshToken, s.config.JWT.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	_ = s.tokenCache.Del(ctx, "2fa:"+claims.UserID)

	return &LoginCompleteOutput{
		UserID:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		AccessToken:      accessToken,
		AccessExpiresIn:  int64(s.config.JWT.AccessTokenTTL.Seconds()),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int64(s.config.JWT.RefreshTokenTTL.Seconds()),
	}, nil
}

func (s *authService) Enable2FA(ctx context.Context, input *Enable2FAInput) (*Enable2FAOutput, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.CheckPassword(input.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	code, err := generate2FACode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate 2FA code: %w", err)
	}

	_, err = s.mailSvc.Send2FACode(ctx, &api.Send2FACodeRequest{
		EmailAddress: user.Email,
		Code:         code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send 2FA code: %w", err)
	}

	err = s.tokenCache.Set(ctx, "enable_2fa:"+user.ID, code, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to store 2FA code: %w", err)
	}

	return &Enable2FAOutput{
		Message: "2FA code sent to email",
	}, nil
}

func (s *authService) Enable2FAComplete(ctx context.Context, input *Enable2FACompleteInput) (*Enable2FACompleteOutput, error) {
	storedCode, err := s.tokenCache.Get(ctx, "enable_2fa:"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("2FA code not found or expired")
	}

	if storedCode != input.Code {
		return nil, fmt.Errorf("invalid 2FA code")
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Is2FAEnabled = true
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	_ = s.tokenCache.Del(ctx, "enable_2fa:"+input.UserID)

	return &Enable2FACompleteOutput{
		Is2FAEnabled: true,
		Message:      "2FA enabled successfully",
	}, nil
}

func (s *authService) Disable2FA(ctx context.Context, input *Disable2FAInput) (*Disable2FAOutput, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.CheckPassword(input.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	if !user.Is2FAEnabled {
		return nil, fmt.Errorf("2FA is not enabled")
	}

	code, err := generate2FACode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	_, err = s.mailSvc.Send2FACode(ctx, &api.Send2FACodeRequest{
		EmailAddress: user.Email,
		Code:         code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	err = s.tokenCache.Set(ctx, "disable_2fa:"+user.ID, code, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to store verification code: %w", err)
	}

	return &Disable2FAOutput{
		Message: "Verification code sent to email",
	}, nil
}

func (s *authService) Disable2FAComplete(ctx context.Context, input *Disable2FACompleteInput) (*Disable2FACompleteOutput, error) {
	storedCode, err := s.tokenCache.Get(ctx, "disable_2fa:"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("verification code not found or expired")
	}

	if storedCode != input.Code {
		return nil, fmt.Errorf("invalid verification code")
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Is2FAEnabled = false
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	_ = s.tokenCache.Del(ctx, "disable_2fa:"+input.UserID)

	return &Disable2FACompleteOutput{
		Is2FAEnabled: false,
		Message:      "2FA disabled successfully",
	}, nil
}

func (s *authService) ChangeEmail(ctx context.Context, input *ChangeEmailInput) (*ChangeEmailOutput, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.CheckPassword(input.CurrentPassword) {
		return nil, fmt.Errorf("invalid password")
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, input.NewEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already in use")
	}

	code, err := generate2FACode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	_, err = s.mailSvc.Send2FACode(ctx, &api.Send2FACodeRequest{
		EmailAddress: input.NewEmail,
		Code:         code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	err = s.tokenCache.Set(ctx, "change_email:"+user.ID, code+":"+input.NewEmail, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to store verification code: %w", err)
	}

	return &ChangeEmailOutput{
		Message: "Verification code sent to new email",
	}, nil
}

func (s *authService) ChangeEmailComplete(ctx context.Context, input *ChangeEmailCompleteInput) (*ChangeEmailCompleteOutput, error) {
	storedData, err := s.tokenCache.Get(ctx, "change_email:"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("verification code not found or expired")
	}

	storedCode, newEmail, ok := splitStoredData(storedData)
	if !ok || storedCode != input.Code {
		return nil, fmt.Errorf("invalid verification code")
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Email = newEmail
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	_ = s.tokenCache.Del(ctx, "change_email:"+input.UserID)
	_ = s.tokenCache.Del(ctx, "refresh:"+user.ID)

	return &ChangeEmailCompleteOutput{
		Email:   newEmail,
		Message: "Email changed successfully",
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.CheckPassword(input.CurrentPassword) {
		return nil, fmt.Errorf("invalid current password")
	}

	code, err := generate2FACode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	_, err = s.mailSvc.Send2FACode(ctx, &api.Send2FACodeRequest{
		EmailAddress: user.Email,
		Code:         code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	err = s.tokenCache.Set(ctx, "change_password:"+user.ID, code+":"+input.NewPassword, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to store verification code: %w", err)
	}

	return &ChangePasswordOutput{
		Message: "Verification code sent to email",
	}, nil
}

func (s *authService) ChangePasswordComplete(ctx context.Context, input *ChangePasswordCompleteInput) (*ChangePasswordCompleteOutput, error) {
	storedData, err := s.tokenCache.Get(ctx, "change_password:"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("verification code not found or expired")
	}

	storedCode, newPassword, ok := splitStoredData(storedData)
	if !ok || storedCode != input.Code {
		return nil, fmt.Errorf("invalid verification code")
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	_ = s.tokenCache.Del(ctx, "change_password:"+input.UserID)
	_ = s.tokenCache.Del(ctx, "refresh:"+user.ID)

	return &ChangePasswordCompleteOutput{
		Message: "Password changed successfully",
	}, nil
}

func (s *authService) ChangeMeta(ctx context.Context, input *ChangeMetaInput) (*ChangeMetaOutput, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Name = input.Name
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &ChangeMetaOutput{
		UserID:  user.ID,
		Name:    user.Name,
		Message: "Profile updated successfully",
	}, nil
}

func splitStoredData(data string) (string, string, bool) {
	for i := 0; i < len(data); i++ {
		if data[i] == ':' {
			return data[:i], data[i+1:], true
		}
	}
	return "", "", false
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
