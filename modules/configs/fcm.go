package configs

import (
	"context"
	"log"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FcmConfig struct {
	Client *messaging.Client
}

var fcmInstance *FcmConfig
var fcmOnce sync.Once

// GetFCM returns a singleton FCM messaging client.
// It reads the service account JSON from the path defined in
// FCM_SERVICE_ACCOUNT_JSON env var, or falls back to
// GOOGLE_APPLICATION_CREDENTIALS if set.
func GetFCM() *messaging.Client {
	fcmOnce.Do(func() {
		log.Println("Initialize FCM connection...")

		ctx := context.Background()

		var opts []option.ClientOption

		// Allow explicit service account path via FCM_SERVICE_ACCOUNT_JSON
		if saPath := os.Getenv("FCM_SERVICE_ACCOUNT_JSON"); saPath != "" {
			opts = append(opts, option.WithCredentialsFile(saPath))
		}

		app, err := firebase.NewApp(ctx, nil, opts...)
		if err != nil {
			log.Panicf("error initializing Firebase app: %v", err)
		}

		client, err := app.Messaging(ctx)
		if err != nil {
			log.Panicf("error getting FCM messaging client: %v", err)
		}

		fcmInstance = &FcmConfig{
			Client: client,
		}

		log.Println("Success Initialize FCM connection...")
	})

	return fcmInstance.Client
}
