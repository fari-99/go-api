package rabbitmq

import (
	"encoding/json"
	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/configs/rabbitmq"
	"net/http"

	"github.com/gin-gonic/gin"
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

	queueSetup := rabbitmq.NewBaseQueue("", "test-queue")
	queueSetup.SetupQueue(nil, nil)
	queueSetup.AddPublisher(&rabbitmq.QueueDeclareConfig{}, &rabbitmq.PublisherConfig{})
	err := queueSetup.Publish("", string(bodyMarshal))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}

func (controller *RabbitMqController) TestBatchPublishQueueAction(ctx *gin.Context) {
	queueSetup := rabbitmq.NewBaseQueue("", "test-queue")
	queueSetup.SetupQueue(nil, nil)
	queueSetup.AddPublisher(&rabbitmq.QueueDeclareConfig{}, &rabbitmq.PublisherConfig{})

	var allMsg []string
	for i := 1; i <= 10; i++ {
		body := map[string]interface{}{
			"testing": "testing",
			"data":    i,
		}

		bodyMarshal, _ := json.Marshal(body)
		allMsg = append(allMsg, string(bodyMarshal))
	}

	allError := queueSetup.BatchPublish("", allMsg)
	if len(allError) > 0 {
		helpers.NewResponse(ctx, http.StatusInternalServerError, allError)
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}
