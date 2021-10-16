package rabbitmq

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/modules/configs"
	"net/http"
)

type RabbitMqController struct {
	*configs.DI
}

func (controller *RabbitMqController) TestPublishQueueAction(ctx *gin.Context) {
	body := map[string]interface{}{
		"testing": "testing",
		"data":    123,
	}

	bodyMarshal, _ := json.Marshal(body)

	queueSetup := controller.Queue.SetQueueName("test-queue")
	queueSetup.AddPublisher(&configs.QueueDeclareConfig{}, &configs.PublisherConfig{})
	err := queueSetup.Publish(string(bodyMarshal))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}

func (controller *RabbitMqController) TestBatchPublishQueueAction(ctx *gin.Context) {
	queueSetup := controller.Queue.SetQueueName("test-queue")
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
		helpers.NewResponse(ctx, http.StatusInternalServerError, allError)
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}
