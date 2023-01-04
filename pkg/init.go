// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package init

import (
	"flag"
	"strconv"

	klog "k8s.io/klog/v2"

	interfaces "github.com/dvonthenen/rabbitmq-patterns/pkg/interfaces"
	manager "github.com/dvonthenen/rabbitmq-patterns/pkg/manager"
)

type LogLevel int64

const (
	LogLevelDefault   LogLevel = iota
	LogLevelErrorOnly          = 1
	LogLevelStandard           = 2
	LogLevelElevated           = 3
	LogLevelFull               = 4
	LogLevelDebug              = 5
	LogLevelTrace              = 6
)

type RabbitInit struct {
	LogLevel      LogLevel
	DebugFilePath string
}

func Init(init RabbitInit) {
	if init.LogLevel == LogLevelDefault {
		init.LogLevel = LogLevelStandard
	}

	klog.InitFlags(nil)
	flag.Set("v", strconv.FormatInt(int64(init.LogLevel), 10))
	if init.DebugFilePath != "" {
		flag.Set("logtostderr", "false")
		flag.Set("log_file", init.DebugFilePath)
	}
	flag.Parse()
}

func New(options interfaces.ManagerOptions) (*interfaces.Manager, error) {
	manager, err := manager.New(manager.ManagerOptions{
		&options,
	})
	if err != nil {
		return nil, err
	}

	var mgrInterface interfaces.Manager
	mgrInterface = manager

	return &mgrInterface, nil
}
