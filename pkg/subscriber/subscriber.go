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

	klog.V(3).Infof("ExchangeDeclare: %s\n", s.GetName())
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

	klog.V(3).Infof("QueueDeclare: %s\n", s.GetName())
	q, err := s.channel.QueueDeclare(
		"",                    // name
		s.options.Durable,     // durable
		s.options.AutoDeleted, // auto-deleted
		s.options.Exclusive,   // exclusive
		s.options.NoWait,      // no-wait
		nil,                   // arguments
	)
	if err != nil {
		klog.V(1).Infof("QueueDeclare %s failed. Err: %v\n", s.GetName(), err)
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return err
	}
	s.queue = &q

	klog.V(3).Infof("QueueBind: %s\n", s.GetName())
	err = s.channel.QueueBind(
		q.Name,         // queue name
		"",             // routing key
		s.options.Name, // exchange
		false,
		nil)
	if err != nil {
		klog.V(1).Infof("QueueBind %s failed. Err: %v\n", s.GetName(), err)
		klog.V(6).Infof("Subscriber.Init LEAVE\n")
		return err
	}

	klog.V(3).Infof("Consume: %s\n", s.GetName())
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
		klog.V(1).Infof("Consume %s failed. Err: %v\n", s.GetName(), err)
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
	klog.V(6).Infof("Subscriber.Retry ENTER\n")
	klog.V(3).Infof("Subscriber.Retry %s called\n", s.GetName())

	// attempt to clean up but continue
	var retErr error
	retErr = nil

	// teardown but keep the channel
	err := s.teardownMinusChannel()
	if err != nil {
		klog.V(1).Infof("teardownMinusChannel failed. Err: %v\n", err)
		retErr = err
	}

	// re-init
	err = s.Init()
	if err == nil {
		klog.V(4).Infof("Subscriber.Retry Succeeded\n")
	} else {
		klog.V(1).Infof("Subscriber.Retry failed. Err: %v\n", err)
		retErr = err
	}
	klog.V(6).Infof("Subscriber.Retry LEAVE\n")

	return retErr
}

func (s *Subscriber) teardownMinusChannel() error {
	s.running = false

	close(s.stopChan)
	<-s.stopChan

	var retErr error
	retErr = nil

	// clean up queue related stuff
	if s.queue != nil {
		err := s.channel.QueueUnbind(s.queue.Name, "", s.options.Name, nil)
		if err != nil {
			klog.V(1).Infof("QueueUnbind %s failed. Err: %v\n", s.queue.Name, err)
			retErr = err
		}

		_, err = s.channel.QueueDelete(s.queue.Name, s.options.IfUnused, s.options.IfEmpty, s.options.NoWait)
		if err != nil {
			klog.V(1).Infof("QueueDelete %s failed. Err: %v\n", s.queue.Name, err)
			retErr = err
		}
		s.queue = nil
	}

	// clean up exchange
	err := s.channel.ExchangeDelete(s.options.Name, s.options.IfUnused, s.options.NoWait)
	if err != nil {
		klog.V(1).Infof("ExchangeDelete %s failed. Err: %v\n", s.GetName(), err)
		retErr = err
	}

	return retErr
}

func (s *Subscriber) Teardown() error {
	klog.V(6).Infof("Subscriber.Teardown ENTER\n")
	klog.V(3).Infof("Subscriber.Teardown %s called\n", s.GetName())

	var retErr error
	retErr = nil

	err := s.teardownMinusChannel()
	if err != nil {
		klog.V(1).Infof("teardownMinusChannel Failed. Err: %v\n", err)
		retErr = err
	}

	if s.channel != nil {
		s.channel.Close()
		s.channel = nil
	}

	if retErr == nil {
		klog.V(4).Infof("Subscriber.Teardown Succeeded\n")
	} else {
		klog.V(1).Infof("Subscriber.Teardown failed. Err: %v\n", retErr)
	}
	klog.V(6).Infof("Subscriber.Teardown LEAVE\n")

	return retErr
}
