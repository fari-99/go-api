package notifications

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/messaging"

	"go-api/modules/configs"
)

// PushNotificationData holds the payload for a single FCM push notification.
type PushNotificationData struct {
	// Token is the FCM registration token of the target device.
	// Either Token OR Topic must be set.
	Token string `json:"token"`

	// Topic sends the notification to all devices subscribed to this topic.
	// Either Token OR Topic must be set.
	Topic string `json:"topic"`

	// Title of the notification.
	Title string `json:"title"`

	// Body is the notification body text.
	Body string `json:"body"`

	// ImageURL is an optional image URL to display in the notification.
	ImageURL string `json:"image_url"`

	// Data is an optional key-value map of custom data sent alongside the notification.
	Data map[string]string `json:"data"`
}

// SendPushNotification sends a single FCM push notification.
// It uses a device registration token when Token is set, or a topic when Topic is set.
func SendPushNotification(data PushNotificationData) error {
	if data.Token == "" && data.Topic == "" {
		return errors.New("either Token or Topic must be provided")
	}

	client := configs.GetFCM()

	msg := buildFCMMessage(data)

	_, err := client.Send(context.Background(), msg)
	if err != nil {
		return err
	}

	return nil
}

// SendMulticastPushNotification sends an FCM notification to multiple device tokens at once.
// Maximum 500 tokens per call (FCM limit).
func SendMulticastPushNotification(tokens []string, data PushNotificationData) error {
	if len(tokens) == 0 {
		return errors.New("at least one token must be provided")
	}

	client := configs.GetFCM()

	multicastMsg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    data.Title,
			Body:     data.Body,
			ImageURL: data.ImageURL,
		},
		Data: data.Data,
	}

	resp, err := client.SendEachForMulticast(context.Background(), multicastMsg)
	if err != nil {
		return err
	}

	if resp.FailureCount > 0 {
		// Collect failed tokens for logging / retry purposes
		var failedTokens []string
		for i, sendResp := range resp.Responses {
			if !sendResp.Success {
				failedTokens = append(failedTokens, tokens[i])
			}
		}

		return errors.New("some tokens failed to receive the notification")
	}

	return nil
}

// buildFCMMessage constructs the FCM Message struct from PushNotificationData.
func buildFCMMessage(data PushNotificationData) *messaging.Message {
	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Title:    data.Title,
			Body:     data.Body,
			ImageURL: data.ImageURL,
		},
		Data: data.Data,
	}

	if data.Token != "" {
		msg.Token = data.Token
	} else {
		msg.Topic = data.Topic
	}

	return msg
}
