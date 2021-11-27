package configs

import (
	"github.com/twilio/twilio-go"
	"os"
	"sync"
)

type twilioConfig struct {
	TwilioRestClient *twilio.RestClient
}

var twilioInstance *twilioConfig
var twilioOnce sync.Once

func GetTwilioRestClient() *twilio.RestClient {
	twilioOnce.Do(func() {
		client := twilio.NewRestClientWithParams(twilio.RestClientParams{
			Username: os.Getenv("TWILIO_ACCOUNT_SID"),
			Password: os.Getenv("TWILIO_AUTH_TOKEN"),
		})

		twilioInstance = &twilioConfig{
			TwilioRestClient: client,
		}
	})

	return twilioInstance.TwilioRestClient
}
