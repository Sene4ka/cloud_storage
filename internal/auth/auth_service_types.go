package auth

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterOutput struct {
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
