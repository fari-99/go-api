package configs

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
	"github.com/redis/go-redis/v9"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"go-api/constant"
)

func WhatsappClient(ctx context.Context, redisClient redis.UniversalClient) *whatsmeow.Client {
	databaseBase := DatabaseBase(PostgresType)
	connUrl := databaseBase.GetConnection()

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New(ctx, strings.ToLower(PostgresType), connUrl+" sslmode=disable", dbLog)
	if err != nil {
		panic(err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// if already connected, then return client
	if client.Store.ID != nil {
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		return client
	}

	// No ID stored, new login
	qrChan, _ := client.GetQRChannel(context.Background())
	err = client.Connect()
	if err != nil {
		panic(err)
	}

	for evt := range qrChan {
		event := evt.Event
		if event == "code" {
			qrCode := evt.Code
			// log.Println("QR code:", qrCode)                               // echo QR Code Value
			if os.Getenv("SHOW_QR_CODE_TERMINAL") == "true" {
				qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout) // print QR Code on terminal
			}

			// save qr code value on redis cache, so it can be used when accidentally logged out
			log.Printf("Please Scan QR-Code on API [GET] /notifications/qr-code/whatsapp")
			_ = redisClient.Set(ctx, constant.QRCodeWhatsapp, qrCode, 0)
		} else {
			log.Printf("Events := %s", event)
		}
	}

	return client
}
