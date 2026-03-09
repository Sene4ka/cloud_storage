package mail

type Send2FACodeInput struct {
	EmailAddress string
	Code         string
}

type Send2FACodeOutput struct {
	Success bool
	Message string
}
