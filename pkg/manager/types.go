// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package manager

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	interfaces "github.com/dvonthenen/rabbitmq-patterns/pkg/interfaces"
	publisher "github.com/dvonthenen/rabbitmq-patterns/pkg/publisher"
	subscriber "github.com/dvonthenen/rabbitmq-patterns/pkg/subscriber"
)

/*
	The one that manages everything
*/
type ManagerOptions struct {
	*interfaces.ManagerOptions
}

type Manager struct {
	// housekeeping
	options ManagerOptions

	publishers  map[string]*publisher.Publisher
	subscribers map[string]*subscriber.Subscriber
	mu          sync.Mutex

	// rabbitmq
	connection *amqp.Connection
}
