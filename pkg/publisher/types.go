// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package publisher

import (
	amqp "github.com/rabbitmq/amqp091-go"

	interfaces "github.com/dvonthenen/rabbitmq-patterns/pkg/interfaces"
)

type PublisherOptions struct {
	*interfaces.PublisherOptions

	Channel *amqp.Channel
}

type Publisher struct {
	options PublisherOptions
	channel *amqp.Channel
}
