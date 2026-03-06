package mail

import (
	"context"

	"github.com/Sene4ka/cloud_storage/internal/api"
)

type Server struct {
	api.UnimplementedMailServiceServer
	service MailService
}

func NewServer(service MailService) *Server {
	return &Server{service: service}
}

func (s *Server) Send2FACode(ctx context.Context, req *api.Send2FACodeRequest) (*api.Send2FACodeResponse, error) {
	out, err := s.service.Send2FACode(ctx, &Send2FACodeInput{
		EmailAddress: req.EmailAddress,
		Code:         req.Code,
	})
	if err != nil {
		return &api.Send2FACodeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &api.Send2FACodeResponse{
		Success: out.Success,
		Message: out.Message,
	}, nil
}
