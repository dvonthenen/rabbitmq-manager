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
		LogLevel: rabbit.LogLevelStandard,
	})

	manager, err := rabbit.New(interfaces.ManagerOptions{
		RabbitURI: "amqp://guest:guest@localhost:5672/",
	})
	if err != nil {
		fmt.Printf("rabbit.New failed. Err: %v\n", err)
		os.Exit(1)
	}

	publisher, err := (*manager).CreatePublisher(interfaces.PublisherOptions{
		Name:        "testing",
		Type:        interfaces.ExchangeTypeFanout,
		AutoDeleted: true,
	})
	if err != nil {
		fmt.Printf("CreatePublisher failed. Err: %v\n", err)
		os.Exit(1)
	}

	err = (*publisher).Init()
	if err != nil {
		fmt.Printf("publisher.Init failed. Err: %v\n", err)
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
		NoAck:       true,
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
	for i := 0; i < 3; i++ {
		time.Sleep(3 * time.Second)
		(*publisher).SendMessage([]byte("hello"))
	}

	// wait for last message to get processed
	time.Sleep(3 * time.Second)

	// restart test
	err = (*manager).Retry()
	if err != nil {
		fmt.Printf("manager.Retry failed. Err: %v\n", err)
	}

	// send message again
	for i := 0; i < 3; i++ {
		time.Sleep(3 * time.Second)
		(*publisher).SendMessage([]byte("hello"))
	}

	// wait for last message to get processed
	time.Sleep(3 * time.Second)

	// teardown
	err = (*manager).Teardown()
	if err != nil {
		fmt.Printf("manager.Teardown failed. Err: %v\n", err)
	}
}
