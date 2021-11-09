package handlers

import (
	"fmt"
	"go-api/modules/configs/rabbitmq"
)

func (base *BaseEventHandler) NotificationsHandler(body rabbitmq.ConsumerHandlerData) {
	fmt.Printf("Handle event [%s]\n", body.EventType)
	var err error
	switch body.EventType {
	case "test":
		err = notificationTest()
	default:
		fmt.Printf("there is no event [%s], ignoring event...\n", body.EventType)
	}

	if err != nil {
		fmt.Printf("error data task [%s] when consumed, err := %s", body.EventType, err.Error())
		return
	}

	fmt.Printf("Sucess handle event [%s]\n", body.EventType)
	return
}

func notificationTest() error {
	panic("im panicked")
}
