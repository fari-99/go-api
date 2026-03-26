package configs

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/redis/go-redis/v9"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"go-api/constant"
)

var (
	waClient     *whatsmeow.Client
	waClientOnce sync.Once
)

// WhatsappClient returns the singleton WhatsApp client.
// It initializes the client on first call and reuses it on subsequent calls.
func WhatsappClient(ctx context.Context, redisClient redis.UniversalClient) *whatsmeow.Client {
	waClientOnce.Do(func() {
		waClient = initWhatsappClient(ctx, redisClient)
	})
	return waClient
}

// InitiateWhatsappLogin triggers a new QR code generation.
// Call this manually (e.g. from your login endpoint) when pairing is needed.
// It is safe to call even if the client is already connected — it will no-op.
func InitiateWhatsappLogin(ctx context.Context, redisClient redis.UniversalClient) error {
	client := WhatsappClient(ctx, redisClient)

	if client.IsConnected() {
		log.Println("[WhatsApp] Already connected, skipping login")
		return nil
	}

	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return err
	}

	if err = client.Connect(); err != nil {
		return err
	}

	// Handle QR events in background — only store the first code, then exit
	go func() {
		for evt := range qrChan {
			switch evt.Event {
			case "code":
				if os.Getenv("SHOW_QR_CODE_TERMINAL") == "true" {
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				}
				// Store the QR code with a 60s TTL — user must scan within this window
				redisClient.Set(ctx, constant.QRCodeWhatsapp, evt.Code, 60*time.Second)
				log.Println("[WhatsApp] QR code ready — scan within 60 seconds")
				return // exit after first code — no looping, no spam

			case "success":
				log.Println("[WhatsApp] QR scanned successfully, session established")
				return

			case "timeout":
				log.Println("[WhatsApp] QR code expired without being scanned")
				return

			default:
				log.Printf("[WhatsApp] QR event: %s", evt.Event)
			}
		}
	}()

	return nil
}

// initWhatsappClient sets up the whatsmeow client with Postgres storage and event handlers.
// If already paired, it connects immediately. Otherwise it waits for a manual login trigger.
func initWhatsappClient(ctx context.Context, redisClient redis.UniversalClient) *whatsmeow.Client {
	databaseBase := DatabaseBase(PostgresType)
	connUrl := databaseBase.GetConnection()

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(ctx, strings.ToLower(PostgresType), connUrl+" sslmode=disable", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected:
			log.Println("[WhatsApp] Connected successfully")
			// Clean up any stale QR code from Redis once connected
			redisClient.Del(ctx, constant.QRCodeWhatsapp)

		case *events.Disconnected:
			log.Println("[WhatsApp] Disconnected from WhatsApp servers")

		case *events.LoggedOut:
			// Session was revoked (e.g. user removed device from WhatsApp app)
			// Reset the singleton so the next call to WhatsappClient re-initializes
			log.Println("[WhatsApp] Logged out — resetting client. Call InitiateWhatsappLogin to re-pair")
			waClientOnce = sync.Once{}
			waClient = nil
		}
	})

	// Already paired — connect immediately, no QR needed
	if client.Store.ID != nil {
		log.Println("[WhatsApp] Existing session found, connecting...")
		if err = client.Connect(); err != nil {
			panic(err)
		}
		return client
	}

	// Not paired — do nothing until InitiateWhatsappLogin is called
	log.Println("[WhatsApp] No session found. Call InitiateWhatsappLogin to generate a QR code")
	return client
}
