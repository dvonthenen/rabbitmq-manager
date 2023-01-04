// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package subscriber

import (
	klog "k8s.io/klog/v2"

	common "github.com/dvonthenen/rabbitmq-patterns/pkg/common"
)

func New(options SubscriberOptions) *Subscriber {
	rabbit := &Subscriber{
		options: options,
		channel: options.Channel,
		handler: options.Handler,
		running: false,
	}
	return rabbit
}

func (s *Subscriber) Init() error {
	klog.V(6).Infof("Subscriber.Init ENTER\n")

	if s.running {
		klog.V(1).Infof("Subscribe already running\n")
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return nil
	}

	klog.V(3).Infof("ExchangeDeclare: %n\n", s.options.Name)
	err := s.channel.ExchangeDeclare(
		s.options.Name, // name
		common.ExchangeTypeToString(s.options.Type), // type
		s.options.Durable,     // durable
		s.options.AutoDeleted, // auto-deleted
		s.options.Internal,    // internal
		s.options.NoWait,      // no-wait
		nil,                   // arguments
	)
	if err != nil {
		klog.V(1).Infof("ExchangeDeclare(%s) failed. Err: %v\n", s.options.Name, err)
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return err
	}

	q, err := s.channel.QueueDeclare(
		"",                    // name
		s.options.Durable,     // durable
		s.options.AutoDeleted, // auto-deleted
		s.options.Exclusive,   // exclusive
		s.options.NoWait,      // no-wait
		nil,                   // arguments
	)
	if err != nil {
		klog.V(1).Infof("QueueDeclare() failed. Err: %v\n", err)
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return err
	}
	s.queue = &q

	klog.V(3).Infof("QueueBind: %n\n", s.options.Name)
	err = s.channel.QueueBind(
		q.Name,         // queue name
		"",             // routing key
		s.options.Name, // exchange
		false,
		nil)
	if err != nil {
		klog.V(1).Infof("QueueBind() failed. Err: %v\n", err)
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return err
	}

	msgs, err := s.channel.Consume(
		s.queue.Name,        // queue
		"",                  // consumer tag
		s.options.NoLocal,   // no local
		s.options.NoAck,     // no ack
		s.options.Exclusive, // exclusive
		s.options.NoWait,    // no wait
		nil,                 // args
	)
	if err != nil {
		klog.V(1).Infof("Consume failed. Err: %v\n", err)
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return err
	}

	klog.V(3).Infof("Subscriber.Init Running message loop...\n")
	s.running = true
	go func() {
		for {
			select {
			default:
				for d := range msgs {
					klog.V(5).Infof(" [x] %s\n", d.Body)

					err := (*s.handler).ProcessMessage(d.Body)
					if err != nil {
						klog.V(1).Infof("ProcessMessage() failed. Err: %v\n", err)
					}
				}
			case <-s.stopChan:
				klog.V(5).Infof("Exiting Subscriber Loop\n")
				return
			}
		}
	}()

	klog.V(4).Infof("Subscriber.Init Succeeded\n")
	klog.V(6).Infof("Subscriber.Init LEAVE\n")

	return nil
}

func (s *Subscriber) Teardown() error {
	klog.V(6).Infof("Subscriber.Teardown ENTER\n")

	s.running = false

	close(s.stopChan)
	<-s.stopChan

	if s.channel != nil {
		s.channel.Close()
		s.channel = nil
	}

	klog.V(4).Infof("Subscriber.Teardown Succeeded\n")
	klog.V(6).Infof("Subscriber.Teardown LEAVE\n")

	return nil
}
