package notifications

import (
	"fmt"
	"os"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers/notifications"
	"go-api/modules/models"
)

type Service interface {
	GetDetail(ctx *gin.Context, id int64) (models.NotificationTemplates, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.NotificationTemplates, *paginator.Pagination, error)
	Create(ctx *gin.Context, request RequestNotificationTemplate) (models.NotificationTemplates, error)
	Update(ctx *gin.Context, id int64, request RequestNotificationTemplate) (models.NotificationTemplates, error)
	Delete(ctx *gin.Context, id int64) error

	QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error)

	// SendManual compiles the given notification template with the target user's data
	// and sends it immediately through the channel matching the template's notification type.
	SendManual(ctx *gin.Context, request RequestSendNotification) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error) {
	return s.repo.QRCodeWhatsapp(ctx)
}

func (s service) GetDetail(ctx *gin.Context, id int64) (models.NotificationTemplates, bool, error) {
	return s.repo.GetDetail(ctx, id)
}

func (s service) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.NotificationTemplates, *paginator.Pagination, error) {
	return s.repo.GetList(ctx, filter)
}

func (s service) Create(ctx *gin.Context, request RequestNotificationTemplate) (models.NotificationTemplates, error) {
	model := models.NotificationTemplates{
		NotificationType: request.NotificationType,
		Action:           request.Action,
		Subject:          request.Subject,
		Body:             request.Body,
		Status:           request.Status,
	}

	return s.repo.Create(ctx, model)
}

func (s service) Update(ctx *gin.Context, id int64, request RequestNotificationTemplate) (models.NotificationTemplates, error) {
	model, notFound, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return models.NotificationTemplates{}, err
	} else if notFound {
		return models.NotificationTemplates{}, fmt.Errorf("notification template not found")
	}

	model.NotificationType = request.NotificationType
	model.Action = request.Action
	model.Subject = request.Subject
	model.Body = request.Body
	model.Status = request.Status

	return s.repo.Update(ctx, model)
}

func (s service) Delete(ctx *gin.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s service) SendManual(ctx *gin.Context, request RequestSendNotification) error {
	template, notFound, err := s.repo.GetDetail(ctx, request.TemplateID)
	if err != nil {
		return err
	} else if notFound {
		return fmt.Errorf("notification template not found")
	} else if template.Status != constant.StatusActive {
		return fmt.Errorf("notification template is not active")
	}

	user, notFound, err := s.repo.GetUserWithSocials(ctx, request.UserID)
	if err != nil {
		return err
	} else if notFound {
		return fmt.Errorf("user not found")
	}

	notificationHelper, err := notifications.NewNotificationHelper([]notifications.NotificationTemplate{
		{
			NotificationType: template.NotificationType,
			Subject:          template.Subject,
			Body:             template.Body,
		},
	})
	if err != nil {
		return err
	}

	compiledTemplates, err := notificationHelper.CompileNotificationTemplate(user)
	if err != nil {
		return err
	}

	compiled, ok := compiledTemplates[template.NotificationType]
	if !ok {
		return fmt.Errorf("failed to compile notification template")
	}

	switch template.NotificationType {
	case constant.NotificationTypeEmail:
		if user.Email == "" {
			return fmt.Errorf("user does not have an email address")
		}

		return notifications.SendEmail(notifications.Email{
			Subject: compiled.Subject,
			Body:    compiled.Body,
			From:    os.Getenv("EMAIL_FROM_DEFAULT"),
			To:      []string{user.Email},
		})
	case constant.NotificationTypeTelegram:
		token, exists := userSocialToken(user, constant.NotificationTypeTelegram)
		if !exists {
			return fmt.Errorf("user does not have telegram linked")
		}

		return notifications.SendTelegram(notifications.TelegramData{
			Message: compiled.Body,
			To:      token,
		})
	case constant.NotificationTypeWhatsapp:
		if user.MobilePhone == "" {
			return fmt.Errorf("user does not have a mobile phone number")
		}

		return notifications.SendWhatsapp(notifications.WhatsappData{
			Message: compiled.Body,
			To:      user.MobilePhone,
		})
	case constant.NotificationTypeSMS:
		if user.MobilePhone == "" {
			return fmt.Errorf("user does not have a mobile phone number")
		}

		return notifications.SendSms(notifications.SmsData{
			Message: compiled.Body,
			To:      user.MobilePhone,
		})
	case constant.NotificationTypePushNotification:
		token, exists := userSocialToken(user, constant.NotificationTypePushNotification)
		if !exists {
			return fmt.Errorf("user does not have a push notification device registered")
		}

		return notifications.SendPushNotification(notifications.PushNotificationData{
			Token: token,
			Title: compiled.Subject,
			Body:  compiled.Body,
		})
	default:
		return fmt.Errorf("unsupported notification type [%d]", template.NotificationType)
	}
}

func userSocialToken(user models.Users, notificationType int8) (string, bool) {
	for _, social := range user.UserSocials {
		if social.NotificationType == notificationType && social.Token != "" {
			return social.Token, true
		}
	}

	return "", false
}
