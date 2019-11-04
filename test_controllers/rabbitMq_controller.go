package test_controllers

import (
	"encoding/json"
	"go-api/configs"
	"go-api/helpers/queue"

	"github.com/kataras/iris"
	"github.com/streadway/amqp"
)

type RabbitMqController struct {
	RabbitMqConnection *amqp.Connection
	RabbitMqChannel    *amqp.Channel
}

func (controller *RabbitMqController) TestPublishQueueAction(ctx iris.Context) {
	baseConnection, err := queue.NewBaseQueue()
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	configQueue := baseConnection.GetDefaultConfigQueue()
	configPublish := baseConnection.GetDefaultConfigPublishQueue()

	configQueue.Name = "test-queue"
	configPublish.Key = configQueue.Name

	body := map[string]interface{}{
		"testing": "testing",
		"data":    123,
	}

	bodyMarshal, _ := json.Marshal(body)

	baseConnection.SetConfigQueue(configQueue).SetConfigPublishQueue(configPublish)
	err = baseConnection.PublishQueue(string(bodyMarshal))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}

func (controller *RabbitMqController) TestBatchPublishQueueAction(ctx iris.Context) {
	baseConnection, err := queue.NewBaseQueue()
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	configQueue := baseConnection.GetDefaultConfigQueue()
	configPublish := baseConnection.GetDefaultConfigPublishQueue()

	configQueue.Name = "test-queue"
	configPublish.Key = configQueue.Name

	var allMsg []string
	for i := 1; i <= 10; i++ {
		body := map[string]interface{}{
			"testing": "testing",
			"data":    i,
		}

		bodyMarshal, _ := json.Marshal(body)
		allMsg = append(allMsg, string(bodyMarshal))
	}

	baseConnection.SetConfigQueue(configQueue).SetConfigPublishQueue(configPublish)
	allError := baseConnection.BatchPublish(allMsg)
	if len(allError) > 0 {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, allError)
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}
