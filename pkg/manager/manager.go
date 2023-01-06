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
	klog.V(6).Infof("Manager.Retry ENTER\n")

	// attempt to restart but return if an error
	var retErr error
	retErr = nil

	for _, publisher := range m.publishers {
		err := publisher.Retry()
		if err != nil {
			klog.V(1).Infof("publisher.Retry %s not found\n", publisher.GetName())
			retErr = err
		}
	}

	for _, subscriber := range m.subscribers {
		err := subscriber.Retry()
		if err != nil {
			klog.V(1).Infof("subscriber.Retry %s not found\n", subscriber.GetName())
			retErr = err
		}
	}

	if retErr == nil {
		klog.V(4).Infof("Manager.Retry Succeeded\n")
	} else {
		klog.V(1).Infof("Manager.Retry failed. Err: %v\n", retErr)
	}
	klog.V(6).Infof("Manager.Retry LEAVE\n")

	return retErr
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

	klog.V(4).Infof("Manager.CreatePublisher %s Succeeded\n", options.Name)
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

	klog.V(4).Infof("Manager.CreateSubscriber %s Succeeded\n", options.Name)
	klog.V(6).Infof("Manager.CreateSubscriber LEAVE\n")

	return &subInterface, nil
}

func (m *Manager) GetPublisherByName(name string) (*interfaces.Publisher, error) {
	klog.V(6).Infof("Manager.GetPublisherByName ENTER\n")
	klog.V(3).Infof("GetPublisherByName: %s\n", name)

	if m.publishers == nil {
		klog.V(1).Infof("publishers map is nil\n")
		klog.V(6).Infof("Manager.GetPublisherByName LEAVE\n")
		return nil, ErrPublisherNotFound
	}

	publisher := m.publishers[name]
	if publisher == nil {
		klog.V(1).Infof("publisher name %s not found\n", name)
		klog.V(6).Infof("Manager.GetPublisherByName LEAVE\n")
		return nil, ErrPublisherNotFound
	}

	var pubInterface interfaces.Publisher
	pubInterface = publisher

	klog.V(4).Infof("GetPublisherByName %s FOUND\n", name)
	klog.V(6).Infof("Manager.GetPublisherByName LEAVE\n")

	return &pubInterface, nil
}

func (m *Manager) GetSubscriberByName(name string) (*interfaces.Subscriber, error) {
	klog.V(6).Infof("Manager.GetSubscriberByName ENTER\n")
	klog.V(3).Infof("GetSubscriberByName: %s\n", name)

	if m.subscribers == nil {
		klog.V(1).Infof("subscribers map is nil\n")
		klog.V(6).Infof("Manager.GetSubscriberByName LEAVE\n")
		return nil, ErrSubscriberNotFound
	}

	subscriber := m.subscribers[name]
	if subscriber == nil {
		klog.V(1).Infof("subscriber name %s not found\n", name)
		klog.V(6).Infof("Manager.GetSubscriberByName LEAVE\n")
		return nil, ErrSubscriberNotFound
	}

	var subInterface interfaces.Subscriber
	subInterface = subscriber

	klog.V(4).Infof("GetSubscriberByName %s FOUND\n", name)
	klog.V(6).Infof("Manager.GetSubscriberByName LEAVE\n")

	return &subInterface, nil
}

func (m *Manager) Init() error {
	klog.V(6).Infof("Manager.Init ENTER\n")

	for msgType, publisher := range m.publishers {
		err := publisher.Init()
		if err == nil {
			klog.V(3).Infof("publisher.Init %s Succeeded\n", msgType)
		} else {
			klog.V(1).Infof("publisher.Init %s failed. Err: %v\n", msgType, err)
			klog.V(6).Infof("Manager.Init LEAVE\n")
			return err
		}
	}

	for msgType, subscriber := range m.subscribers {
		err := subscriber.Init()
		if err == nil {
			klog.V(3).Infof("subscriber.Init %s Succeeded\n", msgType)
		} else {
			klog.V(1).Infof("subscriber.Init %s failed. Err: %v\n", msgType, err)
			klog.V(6).Infof("Manager.Init LEAVE\n")
			return err
		}
	}

	klog.V(4).Infof("Manager.Init Succeeded\n")
	klog.V(6).Infof("Manager.Init LEAVE\n")

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
	klog.V(5).Infof("Data: %s\n", string(data))

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

	klog.V(4).Infof("Manager.PublishMessageByName %s Succeeded\n", name)
	klog.V(6).Infof("Manager.PublishMessageByName LEAVE\n")

	return nil
}

func (m *Manager) DeleteSubscriber(name string) error {
	klog.V(6).Infof("Manager.DeleteSubscriber ENTER\n")
	klog.V(3).Infof("Deleting Subscriber: %s\n", name)

	// find
	if m.subscribers == nil {
		klog.V(1).Infof("subscribers list is nil\n")
		klog.V(6).Infof("Manager.DeleteSubscriber LEAVE\n")
		return ErrPublisherNotFound
	}
	subscriber := m.subscribers[name]
	if subscriber == nil {
		klog.V(1).Infof("subscribers %s not found\n", name)
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

	klog.V(4).Infof("Manager.DeleteSubscriber %s Succeeded\n", name)
	klog.V(6).Infof("Manager.DeleteSubscriber LEAVE\n")

	return nil
}

func (m *Manager) DeletePublisher(name string) error {
	klog.V(6).Infof("Manager.DeletePublisher ENTER\n")
	klog.V(3).Infof("Deleting Publisher: %s\n", name)

	// find
	if m.publishers == nil {
		klog.V(1).Infof("publishers list is nil\n")
		klog.V(6).Infof("Manager.DeleteSubscriber LEAVE\n")
		return ErrPublisherNotFound
	}
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

	klog.V(4).Infof("Manager.DeletePublisher %s Succeeded\n", name)
	klog.V(6).Infof("Manager.DeletePublisher LEAVE\n")

	return nil
}

func (m *Manager) Teardown() error {
	klog.V(6).Infof("Manager.Teardown ENTER\n")

	// attempt to clean everything up but return if an error
	var retErr error
	retErr = nil

	// clean up subs and pubs
	for _, subscriber := range m.subscribers {
		err := subscriber.Teardown()
		if err != nil {
			klog.V(1).Infof("subscriber.Teardown() failed. Err: %v\n", err)
			retErr = err
		}
	}
	for _, publisher := range m.publishers {
		err := publisher.Teardown()
		if err != nil {
			klog.V(1).Infof("subscriber.Teardown() failed. Err: %v\n", err)
			retErr = err
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

	if retErr == nil {
		klog.V(4).Infof("Manager.Teardown Succeeded\n")
	} else {
		klog.V(1).Infof("Manager.Teardown failed. Err: %v\n", retErr)
	}
	klog.V(6).Infof("Manager.Teardown LEAVE\n")

	return retErr
}
