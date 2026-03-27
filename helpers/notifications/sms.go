package notifications

import (
	"errors"
	"log"
	"os"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	"go-api/modules/configs"
)

const SmsTwilio = "twilio"

type SmsData struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

func SendSms(data SmsData) error {
	waClient := os.Getenv("SMS_CLIENT_ENABLED")
	if waClient == "" {
		panic("no SMS_CLIENT_ENABLED environment variable set")
	}

	var err error
	switch waClient {
	case SmsTwilio:
		err = sendSmsTwilio(data)
	default:
		err = errors.New("invalid SMS_CLIENT_ENABLED")
	}

	return err
}

func sendSmsTwilio(data SmsData) error {
	client := configs.GetTwilioRestClient()
	twilioNumber := os.Getenv("TWILIO_NUMBER")
	if twilioNumber == "" {
		panic("no TWILIO_NUMBER environment variable set")
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(data.To) // must be in E.164 format (ex: +628131212121)
	params.SetFrom(twilioNumber)
	params.SetBody(data.Message)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		if resp != nil {
			log.Println("response status:", resp.Status, "response body:", resp.Body)
			log.Println("response error:", resp.ErrorCode, "response error message:", resp.ErrorMessage)
		}

		return err
	}

	return nil
}
