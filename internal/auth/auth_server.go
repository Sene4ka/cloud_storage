package auth

import (
	"context"

	"github.com/Sene4ka/cloud_storage/internal/api"
)

type AuthService interface {
	Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error)
	RegisterComplete(ctx context.Context, input *RegisterCompleteInput) (*RegisterCompleteOutput, error)
	Login(ctx context.Context, input *LoginInput) (*LoginOutput, error)
	LoginComplete(ctx context.Context, input *LoginCompleteInput) (*LoginCompleteOutput, error)
	Refresh(ctx context.Context, input *RefreshInput) (*RefreshOutput, error)
	ValidateToken(ctx context.Context, input *ValidateTokenInput) (*ValidateTokenOutput, error)
	Logout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error)
	Enable2FA(ctx context.Context, input *Enable2FAInput) (*Enable2FAOutput, error)
	Enable2FAComplete(ctx context.Context, input *Enable2FACompleteInput) (*Enable2FACompleteOutput, error)
	Disable2FA(ctx context.Context, input *Disable2FAInput) (*Disable2FAOutput, error)
	Disable2FAComplete(ctx context.Context, input *Disable2FACompleteInput) (*Disable2FACompleteOutput, error)
	ChangeEmail(ctx context.Context, input *ChangeEmailInput) (*ChangeEmailOutput, error)
	ChangeEmailComplete(ctx context.Context, input *ChangeEmailCompleteInput) (*ChangeEmailCompleteOutput, error)
	ChangePassword(ctx context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error)
	ChangePasswordComplete(ctx context.Context, input *ChangePasswordCompleteInput) (*ChangePasswordCompleteOutput, error)
	ChangeMeta(ctx context.Context, input *ChangeMetaInput) (*ChangeMetaOutput, error)
}

type Server struct {
	api.UnimplementedAuthServiceServer
	service AuthService
}

func NewServer(service AuthService) *Server {
	return &Server{service: service}
}

func (s *Server) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	out, err := s.service.Register(ctx, &RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &api.RegisterResponse{
		UserId:  out.UserID,
		Email:   out.Email,
		Name:    out.Name,
		Message: out.Message,
	}, nil
}

func (s *Server) RegisterComplete(ctx context.Context, req *api.RegisterCompleteRequest) (*api.RegisterCompleteResponse, error) {
	out, err := s.service.RegisterComplete(ctx, &RegisterCompleteInput{
		UserID: req.UserId,
		Code:   req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &api.RegisterCompleteResponse{
		UserId:           out.UserID,
		Email:            out.Email,
		Name:             out.Name,
		AccessToken:      out.AccessToken,
		AccessExpiresIn:  out.AccessExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	out, err := s.service.Login(ctx, &LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &api.LoginResponse{
		UserId:           out.UserID,
		Email:            out.Email,
		Name:             out.Name,
		AccessToken:      out.AccessToken,
		AccessExpiresIn:  out.AccessExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
		TempToken:        out.TempToken,
		Requires_2Fa:     out.Requires2FA,
		Message:          out.Message,
	}, nil
}

func (s *Server) LoginComplete(ctx context.Context, req *api.LoginCompleteRequest) (*api.LoginCompleteResponse, error) {
	out, err := s.service.LoginComplete(ctx, &LoginCompleteInput{
		TempToken: req.TempToken,
		Code:      req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &api.LoginCompleteResponse{
		UserId:           out.UserID,
		Email:            out.Email,
		Name:             out.Name,
		AccessToken:      out.AccessToken,
		AccessExpiresIn:  out.AccessExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *api.RefreshRequest) (*api.RefreshResponse, error) {
	out, err := s.service.Refresh(ctx, &RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &api.RefreshResponse{
		AccessToken:      out.AccessToken,
		AccessExpiresIn:  out.AccessExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *api.ValidateTokenRequest) (*api.ValidateTokenResponse, error) {
	out, err := s.service.ValidateToken(ctx, &ValidateTokenInput{
		Token: req.Token,
	})
	if err != nil {
		return &api.ValidateTokenResponse{Valid: false}, nil
	}
	return &api.ValidateTokenResponse{
		Valid:     out.Valid,
		UserId:    out.UserID,
		Email:     out.Email,
		ExpiresIn: out.ExpiresIn,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *api.LogoutRequest) (*api.LogoutResponse, error) {
	out, err := s.service.Logout(ctx, &LogoutInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &api.LogoutResponse{Success: out.Success}, nil
}

func (s *Server) Enable2FA(ctx context.Context, req *api.Enable2FARequest) (*api.Enable2FAResponse, error) {
	out, err := s.service.Enable2FA(ctx, &Enable2FAInput{
		UserID:   req.UserId,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &api.Enable2FAResponse{
		Message: out.Message,
	}, nil
}

func (s *Server) Enable2FAComplete(ctx context.Context, req *api.Enable2FACompleteRequest) (*api.Enable2FACompleteResponse, error) {
	out, err := s.service.Enable2FAComplete(ctx, &Enable2FACompleteInput{
		UserID: req.UserId,
		Code:   req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &api.Enable2FACompleteResponse{
		Is_2FaEnabled: out.Is2FAEnabled,
		Message:       out.Message,
	}, nil
}

func (s *Server) Disable2FA(ctx context.Context, req *api.Disable2FARequest) (*api.Disable2FAResponse, error) {
	out, err := s.service.Disable2FA(ctx, &Disable2FAInput{
		UserID:   req.UserId,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &api.Disable2FAResponse{
		Message: out.Message,
	}, nil
}

func (s *Server) Disable2FAComplete(ctx context.Context, req *api.Disable2FACompleteRequest) (*api.Disable2FACompleteResponse, error) {
	out, err := s.service.Disable2FAComplete(ctx, &Disable2FACompleteInput{
		UserID: req.UserId,
		Code:   req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &api.Disable2FACompleteResponse{
		Is_2FaEnabled: out.Is2FAEnabled,
		Message:       out.Message,
	}, nil
}

func (s *Server) ChangeEmail(ctx context.Context, req *api.ChangeEmailRequest) (*api.ChangeEmailResponse, error) {
	out, err := s.service.ChangeEmail(ctx, &ChangeEmailInput{
		UserID:          req.UserId,
		CurrentPassword: req.CurrentPassword,
		NewEmail:        req.NewEmail,
	})
	if err != nil {
		return nil, err
	}
	return &api.ChangeEmailResponse{
		Message: out.Message,
	}, nil
}

func (s *Server) ChangeEmailComplete(ctx context.Context, req *api.ChangeEmailCompleteRequest) (*api.ChangeEmailCompleteResponse, error) {
	out, err := s.service.ChangeEmailComplete(ctx, &ChangeEmailCompleteInput{
		UserID: req.UserId,
		Code:   req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &api.ChangeEmailCompleteResponse{
		Email:   out.Email,
		Message: out.Message,
	}, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *api.ChangePasswordRequest) (*api.ChangePasswordResponse, error) {
	out, err := s.service.ChangePassword(ctx, &ChangePasswordInput{
		UserID:          req.UserId,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		return nil, err
	}
	return &api.ChangePasswordResponse{
		Message: out.Message,
	}, nil
}

func (s *Server) ChangePasswordComplete(ctx context.Context, req *api.ChangePasswordCompleteRequest) (*api.ChangePasswordCompleteResponse, error) {
	out, err := s.service.ChangePasswordComplete(ctx, &ChangePasswordCompleteInput{
		UserID: req.UserId,
		Code:   req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &api.ChangePasswordCompleteResponse{
		Message: out.Message,
	}, nil
}

func (s *Server) ChangeMeta(ctx context.Context, req *api.ChangeMetaRequest) (*api.ChangeMetaResponse, error) {
	out, err := s.service.ChangeMeta(ctx, &ChangeMetaInput{
		UserID: req.UserId,
		Name:   req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &api.ChangeMetaResponse{
		UserId:  out.UserID,
		Name:    out.Name,
		Message: out.Message,
	}, nil
}
