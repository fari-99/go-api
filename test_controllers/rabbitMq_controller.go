package test_controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go-api/configs"
	"net/http"
)

type RabbitMqController struct {
	QueueSetup *configs.QueueSetup
}

func (controller *RabbitMqController) TestPublishQueueAction(ctx *gin.Context) {
	body := map[string]interface{}{
		"testing": "testing",
		"data":    123,
	}

	bodyMarshal, _ := json.Marshal(body)

	queueSetup := controller.QueueSetup.SetQueueName("test-queue")
	queueSetup.AddPublisher(&configs.QueueDeclareConfig{}, &configs.PublisherConfig{})
	err := queueSetup.Publish(string(bodyMarshal))
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	configs.NewResponse(ctx, http.StatusOK, "yee")
	return
}

func (controller *RabbitMqController) TestBatchPublishQueueAction(ctx *gin.Context) {
	queueSetup := controller.QueueSetup.SetQueueName("test-queue")
	queueSetup.AddPublisher(&configs.QueueDeclareConfig{}, &configs.PublisherConfig{})

	var allMsg []string
	for i := 1; i <= 10; i++ {
		body := map[string]interface{}{
			"testing": "testing",
			"data":    i,
		}

		bodyMarshal, _ := json.Marshal(body)
		allMsg = append(allMsg, string(bodyMarshal))
	}

	allError := queueSetup.BatchPublish(allMsg)
	if len(allError) > 0 {
		configs.NewResponse(ctx, http.StatusInternalServerError, allError)
		return
	}

	configs.NewResponse(ctx, http.StatusOK, "yee")
	return
}
