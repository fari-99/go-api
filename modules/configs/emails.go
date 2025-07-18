package configs

import (
	"log"
	"os"
	"strconv"
	"sync"

	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Dialer *gomail.Dialer
}

var emailSessionInstance *EmailConfig
var emailOnce sync.Once

func GetEmail() *gomail.Dialer {
	emailOnce.Do(func() {
		log.Println("Initialize Email connection...")

		host := os.Getenv("SMTP_SERVER")
		port, _ := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 64)
		username := os.Getenv("SMTP_USERNAME")
		password := os.Getenv("SMTP_PASSWORD")
		dialer := gomail.NewDialer(
			host,
			int(port),
			username,
			password)

		emailSessionInstance = &EmailConfig{
			Dialer: dialer,
		}

		log.Println("Success Initialize Email connection...")
	})

	return emailSessionInstance.Dialer
}
