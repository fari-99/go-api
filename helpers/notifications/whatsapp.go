package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"

	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/configs"
)

const WhatsappTwilio = "twilio"
const WhatsappWhatsmeow = "whatsmeow"

type WhatsappData struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

// whatsappSendLockMinTTL and whatsappSendLockMaxTTL bound how long a single send
// blocks any other send from starting — a random gap in this range so outgoing
// messages are spaced out like a human sending them, instead of firing in a burst.
const whatsappSendLockMinTTL = 500 * time.Millisecond
const whatsappSendLockMaxTTL = 2 * time.Second

// whatsappSendLockMaxWait bounds how long a call will queue behind another
// in-flight send before giving up.
const whatsappSendLockMaxWait = 15 * time.Second
const whatsappSendLockRetryDelay = 100 * time.Millisecond

func SendWhatsapp(data WhatsappData) error {
	waClient := os.Getenv("WHATSAPP_CLIENT_ENABLED")
	if waClient == "" {
		panic("no WHATSAPP_CLIENT_ENABLED environment variable set")
	}

	if err := acquireWhatsappSendLock(); err != nil {
		return fmt.Errorf("failed to acquire whatsapp send lock: %w", err)
	}
	// Deliberately not unlocked: the lock's own TTL (0.5-2s) is what enforces
	// the minimum gap before the next message is allowed to send.

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

// acquireWhatsappSendLock blocks until it can claim a short-lived Redis lock
// shared by all WhatsApp sends, waiting out any lock left by a prior send.
func acquireWhatsappSendLock() error {
	lockTTL := whatsappSendLockMinTTL + rand.N(whatsappSendLockMaxTTL-whatsappSendLockMinTTL)
	maxTries := int(whatsappSendLockMaxWait/whatsappSendLockRetryDelay) + 1

	lockHelper := helpers.RedisLock()
	return lockHelper.Lock(
		constant.RedLockWhatsappSend,
		redsync.WithExpiry(lockTTL),
		redsync.WithTries(maxTries),
		redsync.WithRetryDelay(whatsappSendLockRetryDelay),
	)
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

	// WhatsApp JIDs are digits-only (no leading "+"); a "+" in the user part
	// makes usync/LID lookups on the server time out instead of failing fast.
	targetJid := types.NewJID(normalizePhoneForJid(data.To), types.DefaultUserServer)
	message := &waE2E.Message{
		Conversation: new(data.Message),
	}

	_, err := client.SendMessage(context.Background(), targetJid, message)
	if err != nil {
		return err
	}

	return nil
}

func normalizePhoneForJid(phone string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
}
