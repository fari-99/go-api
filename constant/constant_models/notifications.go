package constant_models

import "go-api/constant"

func GetNotificationTypeLabel() map[int8]string {
	notificationType := make(map[int8]string)

	notificationType[constant.NotificationTypeEmail] = "Email"
	notificationType[constant.NotificationTypeTelegram] = "Telegram"
	notificationType[constant.NotificationTypeWhatsapp] = "WhatsApp"
	notificationType[constant.NotificationTypeSMS] = "SMS"
	notificationType[constant.NotificationTypePushNotification] = "Push Notification"

	return notificationType
}
