package auth

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterOutput struct {
	UserID  string
	Email   string
	Name    string
	Message string
}

type RegisterCompleteInput struct {
	UserID string
	Code   string
}

type RegisterCompleteOutput struct {
	UserID           string
	Email            string
	Name             string
	AccessToken      string
	AccessExpiresIn  int64
	RefreshToken     string
	RefreshExpiresIn int64
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	UserID           string
	Email            string
	Name             string
	AccessToken      string
	AccessExpiresIn  int64
	RefreshToken     string
	RefreshExpiresIn int64
	TempToken        string
	Requires2FA      bool
	Message          string
}

type LoginCompleteInput struct {
	TempToken string
	Code      string
}

type LoginCompleteOutput struct {
	UserID           string
	Email            string
	Name             string
	AccessToken      string
	AccessExpiresIn  int64
	RefreshToken     string
	RefreshExpiresIn int64
}

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	AccessToken      string
	AccessExpiresIn  int64
	RefreshToken     string
	RefreshExpiresIn int64
}

type Enable2FAInput struct {
	UserID   string
	Password string
}

type Enable2FAOutput struct {
	Message string
}

type Enable2FACompleteInput struct {
	UserID string
	Code   string
}

type Enable2FACompleteOutput struct {
	Is2FAEnabled bool
	Message      string
}

type Disable2FAInput struct {
	UserID   string
	Password string
}

type Disable2FAOutput struct {
	Is2FAEnabled bool
	Message      string
}

type Disable2FACompleteInput struct {
	UserID string
	Code   string
}

type Disable2FACompleteOutput struct {
	Is2FAEnabled bool
	Message      string
}

type ChangeEmailInput struct {
	UserID          string
	CurrentPassword string
	NewEmail        string
}

type ChangeEmailOutput struct {
	Message string
}

type ChangeEmailCompleteInput struct {
	UserID string
	Code   string
}

type ChangeEmailCompleteOutput struct {
	Email   string
	Message string
}

type ChangePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordOutput struct {
	Message string
}

type ChangePasswordCompleteInput struct {
	UserID string
	Code   string
}

type ChangePasswordCompleteOutput struct {
	Message string
}

type ChangeMetaInput struct {
	UserID string
	Name   string
}

type ChangeMetaOutput struct {
	UserID  string
	Name    string
	Message string
}

type ValidateTokenInput struct {
	Token string
}

type ValidateTokenOutput struct {
	Valid     bool
	UserID    string
	Email     string
	ExpiresIn int64
}

type LogoutInput struct {
	RefreshToken string
}

type LogoutOutput struct {
	Success bool
}
