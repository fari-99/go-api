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
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"go-api/constant"
)

func init() {
	// Shown in WhatsApp's Linked Devices list in place of the library default "whatsmeow".
	store.SetOSInfo("Chrome", [3]uint32{1, 0, 0})
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
}

var (
	waClient     *whatsmeow.Client
	waClientOnce sync.Once
)

// WhatsappStatus reports whether the WhatsApp client is paired/connected.
type WhatsappStatus struct {
	Connected bool   `json:"connected"`
	LoggedIn  bool   `json:"logged_in"`
	JID       string `json:"jid,omitempty"`
}

// WhatsappConnectionStatus returns the current connection state of the WhatsApp client.
func WhatsappConnectionStatus(ctx context.Context, redisClient redis.UniversalClient) WhatsappStatus {
	client := WhatsappClient(ctx, redisClient)

	// IsConnected() only reflects the websocket handshake, which completes
	// well before pairing does — use IsLoggedIn() so "connected" means the
	// account is actually paired and usable, not just mid-handshake.
	status := WhatsappStatus{
		Connected: client.IsLoggedIn(),
		LoggedIn:  client.IsLoggedIn(),
	}
	if client.Store.ID != nil {
		status.JID = client.Store.ID.String()
	}

	return status
}

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
// It is safe to call even if the client is already paired — it will no-op.
func InitiateWhatsappLogin(ctx context.Context, redisClient redis.UniversalClient) error {
	client := WhatsappClient(ctx, redisClient)

	if client.IsLoggedIn() {
		log.Println("[WhatsApp] Already logged in, skipping login")
		return nil
	}

	// A previous attempt may have left an unauthenticated socket open —
	// GetQRChannel/Connect both refuse to run while already connected.
	if client.IsConnected() {
		client.Disconnect()
	}

	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return err
	}

	if err = client.Connect(); err != nil {
		return err
	}

	// Handle QR events in background. whatsmeow rotates the code roughly
	// every 20s while waiting for a scan, so keep listening and refreshing
	// Redis on every code — not just the first — until success or timeout.
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

			case "success":
				log.Println("[WhatsApp] QR scanned successfully, session established")
				return

			case "timeout":
				log.Println("[WhatsApp] QR code expired without being scanned")
				redisClient.Del(ctx, constant.QRCodeWhatsapp)
				return

			default:
				log.Printf("[WhatsApp] QR event: %s", evt.Event)
			}
		}
	}()

	return nil
}

// WhatsappLogout unlinks the current device (if any) and resets the client
// singleton so a subsequent InitiateWhatsappLogin generates a fresh QR code.
func WhatsappLogout(ctx context.Context, redisClient redis.UniversalClient) error {
	client := WhatsappClient(ctx, redisClient)

	if client.Store.ID != nil {
		if err := client.Logout(ctx); err != nil {
			return err
		}
	} else if client.IsConnected() {
		// Not paired, but a socket may be open from an abandoned QR attempt.
		client.Disconnect()
	} else {
		log.Println("[WhatsApp] Logout requested but no session is linked")
	}

	redisClient.Del(ctx, constant.QRCodeWhatsapp)
	waClientOnce = sync.Once{}
	waClient = nil

	log.Println("[WhatsApp] Logged out — call InitiateWhatsappLogin to re-pair")
	return nil
}

// initWhatsappClient sets up the whatsmeow client with Postgres storage and event handlers.
// If already paired, it connects immediately. Otherwise it waits for a manual login trigger.
func initWhatsappClient(ctx context.Context, redisClient redis.UniversalClient) *whatsmeow.Client {
	databaseBase := DatabaseBase(PostgresType)
	connUrl := databaseBase.GetConnection()

	logLevel := os.Getenv("WHATSAPP_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "WARN"
	}

	dbLog := waLog.Stdout("Database", logLevel, true)
	container, err := sqlstore.New(ctx, strings.ToLower(PostgresType), connUrl+" sslmode=disable", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", logLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected:
			log.Println("[WhatsApp] Connected successfully")
			// Clean up any stale QR code from Redis once connected
			redisClient.Del(ctx, constant.QRCodeWhatsapp)

			// Mark the linked device as online, same as a real WhatsApp Web tab —
			// without this the account looks perpetually offline to WhatsApp's servers.
			if err := client.SendPresence(ctx, types.PresenceAvailable); err != nil {
				log.Printf("[WhatsApp] failed to set presence available: %s", err)
			}

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
