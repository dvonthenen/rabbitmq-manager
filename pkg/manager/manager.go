// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package manager

import (
	klog "k8s.io/klog/v2"

	amqp "github.com/rabbitmq/amqp091-go"

	interfaces "github.com/dvonthenen/rabbitmq-manager/pkg/interfaces"
	publisher "github.com/dvonthenen/rabbitmq-manager/pkg/publisher"
	subscriber "github.com/dvonthenen/rabbitmq-manager/pkg/subscriber"
)

func New(options ManagerOptions) (*Manager, error) {
	conn, err := amqp.Dial(options.RabbitURI)
	if err != nil {
		klog.V(1).Infof("amqp.Dial failed. Err: %v\n", err)
		klog.V(6).Infof("Server.RebuildMessageBus LEAVE\n")
		return nil, err
	}

	rabbit := &Manager{
		subscribers: make(map[string]*subscriber.Subscriber),
		publishers:  make(map[string]*publisher.Publisher),
		connection:  conn,
	}
	return rabbit, nil
}

func (m *Manager) Retry() error {
	for _, publisher := range m.publishers {
		err := publisher.Retry()
		if err != nil {
			klog.V(1).Infof("publisher.Retry %s not found\n", publisher.GetName())
			return err
		}
	}

	for _, subscriber := range m.subscribers {
		err := subscriber.Retry()
		if err != nil {
			klog.V(1).Infof("subscriber.Retry %s not found\n", subscriber.GetName())
			return err
		}
	}

	return nil
}

func (m *Manager) CreatePublisher(options interfaces.PublisherOptions) (*interfaces.Publisher, error) {
	klog.V(6).Infof("Manager.CreatePublisher ENTER\n")

	ch, err := m.connection.Channel()
	if err != nil {
		klog.V(1).Infof("Channel() failed. Err: %v\n", err)
		klog.V(6).Infof("Manager.CreatePublisher LEAVE\n")
		return nil, err
	}

	// setup publisher
	publisherOptions := publisher.PublisherOptions{
		&options,
		ch,
	}
	publisher := publisher.New(publisherOptions)

	err = publisher.Init()
	if err != nil {
		klog.V(1).Infof("Init() failed. Err: %v\n", err)
		klog.V(6).Infof("Manager.CreatePublisher LEAVE\n")
		return nil, err
	}

	m.mu.Lock()
	m.publishers[options.Name] = publisher
	m.mu.Unlock()

	var pubInterface interfaces.Publisher
	pubInterface = publisher

	klog.V(4).Infof("Manager.CreatePublisher(%s) Succeeded\n", options.Name)
	klog.V(6).Infof("Manager.CreatePublisher LEAVE\n")

	return &pubInterface, nil
}

func (m *Manager) CreateSubscriber(options interfaces.SubscriberOptions) (*interfaces.Subscriber, error) {
	klog.V(6).Infof("Manager.CreateSubscriber ENTER\n")

	ch, err := m.connection.Channel()
	if err != nil {
		klog.V(1).Infof("Channel() failed. Err: %v\n", err)
		klog.V(6).Infof("Manager.CreateSubscriber LEAVE\n")
		return nil, err
	}

	// setup subscriber
	subscriberOptions := subscriber.SubscriberOptions{
		&options,
		ch,
	}
	subscriber := subscriber.New(subscriberOptions)

	m.mu.Lock()
	m.subscribers[options.Name] = subscriber
	m.mu.Unlock()

	var subInterface interfaces.Subscriber
	subInterface = subscriber

	klog.V(4).Infof("Manager.CreateSubscriber(%s) Succeeded\n", options.Name)
	klog.V(6).Infof("Manager.CreateSubscriber LEAVE\n")

	return &subInterface, nil
}

func (m *Manager) GetPublisherByName(name string) (*interfaces.Publisher, error) {
	if m.publishers == nil {
		return nil, ErrPublisherNotFound
	}

	publisher := m.publishers[name]
	if publisher == nil {
		return nil, ErrPublisherNotFound
	}

	var pubInterface interfaces.Publisher
	pubInterface = publisher

	return &pubInterface, nil
}

func (m *Manager) GetSubscriberByName(name string) (*interfaces.Subscriber, error) {
	if m.subscribers == nil {
		return nil, ErrSubscriberNotFound
	}

	subscriber := m.subscribers[name]
	if subscriber == nil {
		return nil, ErrSubscriberNotFound
	}

	var subInterface interfaces.Subscriber
	subInterface = subscriber

	return &subInterface, nil
}

func (m *Manager) Start() error {
	klog.V(6).Infof("Manager.Start ENTER\n")

	for msgType, subscriber := range m.subscribers {
		err := subscriber.Init()
		if err == nil {
			klog.V(3).Infof("subscriber.Start(%s) Succeeded\n", msgType)
		} else {
			klog.V(1).Infof("subscriber.Start() failed. Err: %v\n", err)
			klog.V(6).Infof("Manager.Start LEAVE\n")
		}
	}

	klog.V(4).Infof("Manager.Start Succeeded\n")
	klog.V(6).Infof("Manager.Start LEAVE\n")

	return nil
}

func (m *Manager) Stop() error {
	klog.V(6).Infof("Manager.Stop ENTER\n")

	for _, subscriber := range m.subscribers {
		err := subscriber.Teardown()
		if err != nil {
			klog.V(1).Infof("subscriber.Teardown() failed. Err: %v\n", err)
		}
	}

	klog.V(4).Infof("Manager.Stop Succeeded\n")
	klog.V(6).Infof("Manager.Stop LEAVE\n")

	return nil
}

func (m *Manager) PublishMessageByName(name string, data []byte) error {
	klog.V(6).Infof("Manager.PublishMessageByName ENTER\n")

	klog.V(3).Infof("Publishing to: %s\n", name)
	klog.V(3).Infof("Data: %s\n", string(data))
	publisher := m.publishers[name]
	if publisher == nil {
		klog.V(1).Infof("Publisher %s not found\n", name)
		klog.V(6).Infof("Manager.PublishMessageByName LEAVE\n")
		return ErrPublisherNotFound
	}

	err := publisher.SendMessage(data)
	if err != nil {
		klog.V(1).Infof("SendMessage() failed. Err: %v\n", err)
		klog.V(6).Infof("Manager.PublishMessageByName LEAVE\n")
		return err
	}

	klog.V(4).Infof("Manager.PublishMessageByName Succeeded\n")
	klog.V(6).Infof("Manager.PublishMessageByName LEAVE\n")

	return nil
}

func (m *Manager) DeleteSubscriber(name string) error {
	klog.V(6).Infof("Manager.DeleteSubscriber ENTER\n")

	// find
	klog.V(3).Infof("Deleting Subscriber: %s\n", name)
	subscriber := m.subscribers[name]
	if subscriber == nil {
		klog.V(1).Infof("Subscriber %s not found\n", name)
		klog.V(6).Infof("Manager.DeleteSubscriber LEAVE\n")
		return ErrPublisherNotFound
	}

	// clean up
	err := subscriber.Teardown()
	if err != nil {
		klog.V(1).Infof("Subscriber.Teardown failed. Err: %v\n", err)
	}

	m.mu.Lock()
	delete(m.subscribers, name)
	m.mu.Unlock()

	klog.V(4).Infof("Manager.DeleteSubscriber Succeeded\n")
	klog.V(6).Infof("Manager.DeleteSubscriber LEAVE\n")

	return nil
}

func (m *Manager) DeletePublisher(name string) error {
	klog.V(6).Infof("Manager.DeletePublisher ENTER\n")

	// find
	klog.V(3).Infof("Deleting Publisher: %s\n", name)
	publisher := m.publishers[name]
	if publisher == nil {
		klog.V(1).Infof("Publisher %s not found\n", name)
		klog.V(6).Infof("Manager.DeletePublisher LEAVE\n")
		return ErrPublisherNotFound
	}

	// clean up
	err := publisher.Teardown()
	if err != nil {
		klog.V(1).Infof("Publisher.Teardown failed. Err: %v\n", err)
	}

	m.mu.Lock()
	delete(m.publishers, name)
	m.mu.Unlock()

	klog.V(4).Infof("Manager.DeletePublisher Succeeded\n")
	klog.V(6).Infof("Manager.DeletePublisher LEAVE\n")

	return nil
}

func (m *Manager) Teardown() error {
	klog.V(6).Infof("Manager.Teardown ENTER\n")

	// clean up subs and pubs
	for _, subscriber := range m.subscribers {
		err := subscriber.Teardown()
		if err != nil {
			klog.V(1).Infof("subscriber.Teardown() failed. Err: %v\n", err)
		}
	}
	for _, publisher := range m.publishers {
		err := publisher.Teardown()
		if err != nil {
			klog.V(1).Infof("subscriber.Teardown() failed. Err: %v\n", err)
		}
	}

	m.mu.Lock()
	m.subscribers = make(map[string]*subscriber.Subscriber)
	m.publishers = make(map[string]*publisher.Publisher)
	m.mu.Unlock()

	// clean up rabbitmq
	if m.connection != nil {
		m.connection.Close()
		m.connection = nil
	}

	klog.V(4).Infof("Manager.Teardown Succeeded\n")
	klog.V(6).Infof("Manager.Teardown LEAVE\n")

	return nil
}
