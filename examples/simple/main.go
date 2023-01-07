package main

import (
	"fmt"
	"os"
	"time"

	rabbit "github.com/dvonthenen/rabbitmq-manager/pkg"
	interfaces "github.com/dvonthenen/rabbitmq-manager/pkg/interfaces"
)

type MyHandler struct{}

func (my MyHandler) ProcessMessage(byData []byte) error {
	fmt.Printf("MSG:\n%s\n\n", string(byData))
	return nil
}

func main() {
	rabbit.Init(rabbit.RabbitInit{
		LogLevel: rabbit.LogLevelStandard, // LogLevelStandard / LogLevelTrace
	})

	manager, err := rabbit.New(interfaces.ManagerOptions{
		RabbitURI: "amqp://guest:guest@localhost:5672/",
		// DeleteWarnings: true,
	})
	if err != nil {
		fmt.Printf("rabbit.New failed. Err: %v\n", err)
		os.Exit(1)
	}

	publisher, err := (*manager).CreatePublisher(interfaces.PublisherOptions{
		Name:        "testing",
		Type:        interfaces.ExchangeTypeFanout,
		AutoDeleted: true,
		IfUnused:    true,
	})
	if err != nil {
		fmt.Printf("CreatePublisher failed. Err: %v\n", err)
		os.Exit(1)
	}

	// object
	thisHandler := MyHandler{}

	// interface reference
	var myHandler interfaces.RabbitMessageHandler
	myHandler = thisHandler

	_, err = (*manager).CreateSubscriber(interfaces.SubscriberOptions{
		Name:        "testing",
		Type:        interfaces.ExchangeTypeFanout,
		AutoDeleted: true,
		IfUnused:    true,
		Handler:     &myHandler,
	})
	if err != nil {
		fmt.Printf("CreateSubscriber failed. Err: %v\n", err)
		os.Exit(1)
	}

	err = (*manager).Init()
	if err != nil {
		fmt.Printf("manager.Init failed. Err: %v\n", err)
		os.Exit(1)
	}

	// send message
	(*publisher).SendMessage([]byte("hello"))

	// wait for message to get processed
	time.Sleep(3 * time.Second)

	// teardown
	err = (*manager).Teardown()
	if err != nil {
		fmt.Printf("manager.Teardown failed. Err: %v\n", err)
	}
}
