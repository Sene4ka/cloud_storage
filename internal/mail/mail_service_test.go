package mail

import (
	"context"
	"errors"
	"testing"

	"github.com/Sene4ka/cloud_storage/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/gomail.v2"
)

type MockSMTPSender struct {
	mock.Mock
}

func (m *MockSMTPSender) Send(msg *gomail.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestMailService_Send2FACode_Success(t *testing.T) {
	t.Parallel()

	mockSender := new(MockSMTPSender)
	config := &configs.Config{
		SMTP: configs.SMTPConfig{
			Host:         "smtp.example.com",
			Port:         "587",
			EmailAddress: "noreply@example.com",
			Password:     "password",
		},
	}

	svc := NewMailService(config, mockSender)

	mockSender.On("Send", mock.Anything).Return(nil)

	input := &Send2FACodeInput{
		EmailAddress: "user@example.com",
		Code:         "123456",
	}

	output, err := svc.Send2FACode(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Success)
	assert.Contains(t, output.Message, "успешно")
	mockSender.AssertExpectations(t)
}

func TestMailService_Send2FACode_SMTPError(t *testing.T) {
	t.Parallel()

	mockSender := new(MockSMTPSender)
	config := &configs.Config{
		SMTP: configs.SMTPConfig{
			Host:         "smtp.example.com",
			Port:         "587",
			EmailAddress: "noreply@example.com",
			Password:     "password",
		},
	}

	svc := NewMailService(config, mockSender)

	mockSender.On("Send", mock.Anything).Return(errors.New("connection refused"))

	input := &Send2FACodeInput{
		EmailAddress: "user@example.com",
		Code:         "123456",
	}

	output, err := svc.Send2FACode(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.False(t, output.Success)
	assert.Contains(t, output.Message, "SMTP Error")
	assert.Contains(t, output.Message, "connection refused")
	mockSender.AssertExpectations(t)
}

func TestMailService_Send2FACode_InvalidEmail(t *testing.T) {
	t.Parallel()

	mockSender := new(MockSMTPSender)
	config := &configs.Config{
		SMTP: configs.SMTPConfig{
			Host:         "smtp.example.com",
			Port:         "587",
			EmailAddress: "noreply@example.com",
			Password:     "password",
		},
	}

	svc := NewMailService(config, mockSender)

	mockSender.On("Send", mock.Anything).Return(errors.New("invalid recipient"))

	input := &Send2FACodeInput{
		EmailAddress: "invalid-email",
		Code:         "123456",
	}

	output, err := svc.Send2FACode(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.False(t, output.Success)
	mockSender.AssertExpectations(t)
}
