package auth

import (
	"context"

	"github.com/Sene4ka/cloud_storage/internal/api"
)

type AuthService interface {
	Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input *LoginInput) (*LoginOutput, error)
	Refresh(ctx context.Context, input *RefreshInput) (*RefreshOutput, error)
	ValidateToken(ctx context.Context, input *ValidateTokenInput) (*ValidateTokenOutput, error)
	Logout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error)
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
