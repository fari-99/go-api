package notifications

import (
	"fmt"

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

	headers := make(map[string][]string)
	if email.From != "" {
		return fmt.Errorf("invalid email from address")
	}

	headers["From"] = []string{email.From}

	if len(email.To) == 0 {
		return fmt.Errorf("invalid email to address")
	}

	headers["To"] = email.To

	if len(email.Cc) > 0 {
		headers["Cc"] = email.Cc
	}

	if len(email.Bcc) > 0 {
		headers["Bcc"] = email.Bcc
	}

	headers["Subject"] = []string{email.Subject}

	m := gomail.NewMessage()
	m.SetHeaders(headers)

	m.SetBody("text/html", email.Body)

	if err := dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
