package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	rabbit "github.com/dvonthenen/rabbitmq-patterns/pkg"
	interfaces "github.com/dvonthenen/rabbitmq-patterns/pkg/interfaces"
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
		Durable:     true,
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

	subscriber, err := (*manager).CreateSubscriber(interfaces.SubscriberOptions{
		Name:        "testing",
		Type:        interfaces.ExchangeTypeFanout,
		Durable:     true,
		AutoDeleted: true,
		NoAck:       true,
		Handler:     &myHandler,
	})
	if err != nil {
		fmt.Printf("CreateSubscriber failed. Err: %v\n", err)
		os.Exit(1)
	}

	err = (*subscriber).Init()
	if err != nil {
		fmt.Printf("subscriber.Init failed. Err: %v\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			time.Sleep(3 * time.Second)
			(*publisher).SendMessage([]byte("hello"))
		}
	}()

	fmt.Print("Press ENTER to exit!\n\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}
