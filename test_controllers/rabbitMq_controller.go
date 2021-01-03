package test_controllers

import (
	"encoding/json"
	"go-api/configs"

	"github.com/kataras/iris/v12"
)

type RabbitMqController struct {
	QueueSetup *configs.QueueSetup
}

func (controller *RabbitMqController) TestPublishQueueAction(ctx iris.Context) {
	body := map[string]interface{}{
		"testing": "testing",
		"data":    123,
	}

	bodyMarshal, _ := json.Marshal(body)

	queueSetup := controller.QueueSetup.SetQueueName("test-queue")
	queueSetup.AddPublisher(&configs.QueueDeclareConfig{}, &configs.PublisherConfig{})
	err := queueSetup.Publish(string(bodyMarshal))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusOK, err.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}

func (controller *RabbitMqController) TestBatchPublishQueueAction(ctx iris.Context) {
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

	//allError := queueSetup.BatchPublish(allMsg)
	//if len(allError) > 0 {
	//	_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, allError)
	//	return
	//}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}
