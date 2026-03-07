package mail

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Sene4ka/cloud_storage/configs"
	"gopkg.in/gomail.v2"
)

type SMTPSender interface {
	Send(msg *gomail.Message) error
}

type mailService struct {
	config     *configs.Config
	smtpSender SMTPSender
}

func NewMailService(config *configs.Config, smtpSender SMTPSender) *mailService {
	return &mailService{
		config:     config,
		smtpSender: smtpSender,
	}
}

func NewMailServiceWithDialer(config *configs.Config) (*mailService, error) {
	port, err := strconv.Atoi(config.SMTP.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP port: %w", err)
	}
	dialer := gomail.NewDialer(config.SMTP.Host, port, config.SMTP.EmailAddress, config.SMTP.Password)
	smtpSender := NewGomailAdapter(dialer)
	return NewMailService(config, smtpSender), nil
}

func (s *mailService) Send2FACode(ctx context.Context, input *Send2FACodeInput) (*Send2FACodeOutput, error) {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.SMTP.EmailAddress)
	m.SetHeader("To", input.EmailAddress)
	m.SetHeader("Subject", "Ваш код подтверждения Cloud Storage")

	htmlBody := fmt.Sprintf(`
        <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto;">
            <h2 style="color: #333;">Код подтверждения</h2>
            <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); 
                        color: white; padding: 30px; text-align: center; border-radius: 12px; 
                        font-size: 36px; font-weight: bold; letter-spacing: 6px; 
                        box-shadow: 0 10px 30px rgba(0,0,0,0.2);">
                %s
            </div>
            <p style="color: #666; margin-top: 20px;">Этот код действителен <strong>5 минут</strong>.</p>
            <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
            <p style="color: #999; font-size: 14px;">Cloud Storage &bull; %s</p>
        </div>`, input.Code, s.config.SMTP.EmailAddress)

	m.SetBody("text/html", htmlBody)
	m.SetBody("text/plain", fmt.Sprintf("Ваш код 2FA: %s (действителен 5 минут)", input.Code))

	if err := s.smtpSender.Send(m); err != nil {
		return &Send2FACodeOutput{
			Success: false,
			Message: fmt.Sprintf("SMTP Error: %v", err),
		}, nil
	}

	return &Send2FACodeOutput{
		Success: true,
		Message: "Код успешно отправлен",
	}, nil
}
