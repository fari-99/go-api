package notifications

import (
	"go-api/modules/configs"

	"gopkg.in/gomail.v2"
)

type Email struct {
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
}

func SendEmail(email Email) error {
	dialer := configs.GetEmail()

	m := gomail.NewMessage()
	m.SetHeaders(map[string][]string{
		"From":    {email.From},
		"To":      email.To,
		"Cc":      email.Cc,
		"Bcc":     email.Bcc,
		"Subject": {email.Subject},
	})

	m.SetBody("text/html", email.Body)

	if err := dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
