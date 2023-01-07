// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package interfaces

/*
	Configuration Options for Rabbit Subscriber/Publisher
*/
type ManagerOptions struct {
	RabbitURI      string
	DeleteWarnings bool
}

type PublisherOptions struct {
	Name string
	Type ExchangeType

	// init
	Durable     bool
	AutoDeleted bool
	Internal    bool

	// teardown
	IfUnused       bool
	DeleteWarnings bool

	// common
	NoWait bool
}

type SubscriberOptions struct {
	Name    string
	Type    ExchangeType
	Handler *RabbitMessageHandler

	// init
	Durable     bool
	AutoDeleted bool
	Internal    bool
	Exclusive   bool
	NoLocal     bool
	NoAck       bool

	// teardown
	IfUnused       bool
	IfEmpty        bool
	DeleteWarnings bool

	// common
	NoWait bool
}

/*
	Object interfaces
*/
type Publisher interface {
	GetName() string
	Init() error
	Retry() error
	SendMessage([]byte) error
	Teardown() error
}

type Subscriber interface {
	GetName() string
	Init() error
	Retry() error
	Teardown() error
}

/*
	Each Subscriber implements this message handler which serves as a callback.
	This function is called when the Rabbit Subscriber receives a message from a named Subscriber
*/
type RabbitMessageHandler interface {
	ProcessMessage(byData []byte) error
}

/*
	Interface to the Rabbit Manager which keeps track of all Publishers and Subscribers
	for a given instance
*/
type Manager interface {
	Init() error
	Retry() error
	CreatePublisher(options PublisherOptions) (*Publisher, error)
	CreateSubscriber(options SubscriberOptions) (*Subscriber, error)
	GetPublisherByName(name string) (*Publisher, error)
	GetSubscriberByName(name string) (*Subscriber, error)
	PublishMessageByName(name string, data []byte) error
	DeletePublisher(name string) error
	DeleteSubscriber(name string) error
	Teardown() error
}
