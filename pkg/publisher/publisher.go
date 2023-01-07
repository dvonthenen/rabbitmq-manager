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

	klog.V(3).Infof("ExchangeDeclare: %s\n", p.GetName())
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
	klog.V(6).Infof("Publisher.Retry ENTER\n")
	klog.V(3).Infof("Publisher.Retry %s called\n", p.GetName())

	err := p.Init()
	if err == nil {
		klog.V(4).Infof("Publisher.Retry Succeeded\n")
	} else {
		klog.V(1).Infof("Publisher.Retry failed. Err: %v\n", err)
	}
	klog.V(6).Infof("Publisher.Retry LEAVE\n")

	return err
}

func (p *Publisher) SendMessage(data []byte) error {
	klog.V(6).Infof("Publisher.SendMessage ENTER\n")
	klog.V(3).Infof("Publishing to: %s\n", p.options.Name)
	klog.V(4).Infof("Data: %s\n", string(data))

	ctx := context.Background()
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
		klog.V(6).Infof("Publisher.SendMessage LEAVE\n")
		return err
	}

	klog.V(4).Infof("Publisher.SendMessage %s succeeded\n%s\n", p.GetName(), string(data))
	klog.V(6).Infof("Publisher.SendMessage LEAVE\n")

	return nil
}

func (p *Publisher) teardownMinusChannel() error {
	var retErr error
	retErr = nil

	// clean up exchange
	err := p.channel.ExchangeDelete(p.options.Name, p.options.IfUnused, p.options.NoWait)
	if err != nil {
		publishError, ok := err.(*amqp.Error)
		if ok {
			if publishError.Code != 504 && publishError.Code != 406 {
				klog.V(1).Infof("ExchangeDelete %s failed. Err: %v\n", p.GetName(), err)
				retErr = err
			} else if p.options.DeleteWarnings {
				klog.V(1).Infof("ExchangeDelete %s failed. Err: %v\n", p.GetName(), err)
				retErr = err
			}
		} else {
			retErr = common.ErrUnresolvedRabbitError
		}
	}

	return retErr
}

func (p *Publisher) Teardown() error {
	klog.V(6).Infof("Publisher.Teardown ENTER\n")
	klog.V(3).Infof("Publisher.Teardown %s called\n", p.GetName())

	var retErr error
	retErr = nil

	err := p.teardownMinusChannel()
	if err != nil {
		klog.V(1).Infof("teardownMinusChannel Failed. Err: %v\n", err)
		retErr = err
	}

	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}

	if retErr == nil {
		klog.V(4).Infof("Publisher.Teardown Succeeded\n")
	} else {
		klog.V(1).Infof("Publisher.Teardown failed. Err: %v\n", retErr)
	}
	klog.V(6).Infof("Publisher.Teardown LEAVE\n")

	return retErr
}
