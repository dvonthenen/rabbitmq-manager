// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package publisher

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	klog "k8s.io/klog/v2"

	common "github.com/dvonthenen/rabbitmq-manager/pkg/common"
)

func New(options PublisherOptions) *Publisher {
	rabbit := &Publisher{
		options: options,
		channel: options.Channel,
	}
	return rabbit
}

func (p *Publisher) GetName() string {
	return p.options.Name
}

func (p *Publisher) Init() error {
	klog.V(6).Infof("Publisher.Init ENTER\n")

	err := p.channel.ExchangeDeclare(
		p.options.Name, // name
		common.ExchangeTypeToString(p.options.Type), // type
		p.options.Durable,     // durable
		p.options.AutoDeleted, // auto-deleted
		p.options.Internal,    // internal
		p.options.NoWait,      // no-wait
		nil,                   // arguments
	)
	if err != nil {
		klog.V(1).Infof("ExchangeDeclare failed. Err: %v\n", err)
		klog.V(6).Infof("Publisher.Init LEAVE\n")
		return err
	}

	klog.V(4).Infof("Publisher.Init Succeeded\n")
	klog.V(6).Infof("Publisher.Init LEAVE\n")

	return nil
}

func (p *Publisher) Retry() error {
	return p.Init()
}

func (p *Publisher) SendMessage(data []byte) error {
	klog.V(6).Infof("Publisher.SendMessage ENTER\n")

	ctx := context.Background()

	klog.V(3).Infof("Publishing to: %s\n", p.options.Name)
	klog.V(3).Infof("Data: %s\n", string(data))
	err := p.channel.PublishWithContext(ctx,
		p.options.Name, // exchange
		"",             // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})
	if err != nil {
		klog.V(1).Infof("PublishWithContext failed. Err: %v\n", err)
		return err
	}

	klog.V(3).Infof("Publisher.SendMessage succeeded\n%s\n", string(data))
	klog.V(6).Infof("Publisher.SendMessage LEAVE\n")

	return nil
}

func (p *Publisher) teardownMinusChannel() error {
	// clean up exchange
	_ = p.channel.ExchangeDelete(p.options.Name, p.options.IfUnused, p.options.NoWait)

	return nil
}

func (p *Publisher) Teardown() error {
	klog.V(6).Infof("Publisher.Teardown ENTER\n")

	err := p.teardownMinusChannel()
	if err != nil {
		return err
	}

	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}

	klog.V(4).Infof("Publisher.Teardown Succeeded\n")
	klog.V(6).Infof("Publisher.Teardown LEAVE\n")

	return nil
}
