// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package subscriber

import (
	klog "k8s.io/klog/v2"

	common "github.com/dvonthenen/rabbitmq-manager/pkg/common"
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

func (s *Subscriber) GetName() string {
	return s.options.Name
}

func (s *Subscriber) Init() error {
	klog.V(6).Infof("Subscriber.Init ENTER\n")

	if s.running {
		klog.V(1).Infof("Subscribe already running\n")
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return nil
	}

	klog.V(3).Infof("ExchangeDeclare: %s\n", s.options.Name)
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
		klog.V(1).Infof("ExchangeDeclare %s failed. Err: %v\n", s.options.Name, err)
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

	klog.V(3).Infof("QueueBind: %s\n", s.options.Name)
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
	s.stopChan = make(chan struct{})
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

func (s *Subscriber) Retry() error {
	// teardown but keep the channel
	err := s.teardownMinusChannel()
	if err != nil {
		return err
	}

	// re-init
	return s.Init()
}

func (s *Subscriber) teardownMinusChannel() error {
	s.running = false

	close(s.stopChan)
	<-s.stopChan

	// clean up queue related stuff
	if s.queue != nil {
		err := s.channel.QueueUnbind(s.queue.Name, "", s.options.Name, nil)
		if err != nil {
			klog.V(1).Infof("QueueUnbind failed. Err: %v\n", err)
		}

		_, err = s.channel.QueueDelete(s.queue.Name, s.options.IfUnused, s.options.IfEmpty, s.options.NoWait)
		if err != nil {
			klog.V(1).Infof("QueueDelete failed. Err: %v\n", err)
		}
		s.queue = nil
	}

	// clean up exchange
	_ = s.channel.ExchangeDelete(s.options.Name, s.options.IfUnused, s.options.NoWait)

	return nil
}

func (s *Subscriber) Teardown() error {
	klog.V(6).Infof("Subscriber.Teardown ENTER\n")

	err := s.teardownMinusChannel()
	if err != nil {
		return err
	}

	if s.channel != nil {
		s.channel.Close()
		s.channel = nil
	}

	klog.V(4).Infof("Subscriber.Teardown Succeeded\n")
	klog.V(6).Infof("Subscriber.Teardown LEAVE\n")

	return nil
}
