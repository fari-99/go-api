package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"go-api/modules/configs"
)

const WhatsappTwilio = "twilio"
const WhatsappWhatsmeow = "whatsmeow"

type WhatsappData struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

func SendWhatsapp(data WhatsappData) error {
	waClient := os.Getenv("WHATSAPP_CLIENT_ENABLED")
	if waClient == "" {
		panic("no WHATSAPP_CLIENT_ENABLED environment variable set")
	}

	var err error
	switch waClient {
	case WhatsappTwilio:
		err = sendWhatsappTwilio(data)
	case WhatsappWhatsmeow:
		err = sendWhatsappWhatsmeow(data)
	default:
		err = errors.New("invalid WHATSAPP_CLIENT_ENABLED")
	}

	return err
}

func sendWhatsappTwilio(data WhatsappData) error {
	client := configs.GetTwilioRestClient()
	twilioNumber := os.Getenv("TWILIO_NUMBER")
	if twilioNumber == "" {
		panic("no TWILIO_NUMBER environment variable set")
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(fmt.Sprintf("whatsapp:%s", data.To))
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

func sendWhatsappWhatsmeow(data WhatsappData) error {
	redisClient := configs.GetRedis(configs.REDIS_SESSION_PREFIX)
	client := configs.WhatsappClient(context.Background(), redisClient)

	targetJid := types.NewJID(data.To, types.DefaultUserServer)
	message := &waE2E.Message{
		Conversation: proto.String(data.Message),
	}

	_, err := client.SendMessage(context.Background(), targetJid, message)
	if err != nil {
		return err
	}

	return nil
}
