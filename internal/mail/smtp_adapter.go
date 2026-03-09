package mail

import "gopkg.in/gomail.v2"

type GomailAdapter struct {
	dialer *gomail.Dialer
}

func NewGomailAdapter(dialer *gomail.Dialer) *GomailAdapter {
	return &GomailAdapter{dialer: dialer}
}

func (a *GomailAdapter) Send(msg *gomail.Message) error {
	return a.dialer.DialAndSend(msg)
}
