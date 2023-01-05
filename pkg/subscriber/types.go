// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package subscriber

import (
	amqp "github.com/rabbitmq/amqp091-go"

	interfaces "github.com/dvonthenen/rabbitmq-manager/pkg/interfaces"
)

type SubscriberOptions struct {
	*interfaces.SubscriberOptions

	Channel *amqp.Channel
}

type Subscriber struct {
	options  SubscriberOptions
	channel  *amqp.Channel
	queue    *amqp.Queue
	stopChan chan struct{}
	handler  *interfaces.RabbitMessageHandler
	running  bool
}
