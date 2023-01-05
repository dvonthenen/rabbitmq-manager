// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package common

import (
	interfaces "github.com/dvonthenen/rabbitmq-manager/pkg/interfaces"
)

func ExchangeTypeToString(exchangeType interfaces.ExchangeType) string {
	switch exchangeType {
	case interfaces.ExchangeTypeDirect:
		return ExchangeDirect
	case interfaces.ExchangeTypeFanout:
		return ExchangeFanout
	case interfaces.ExchangeTypeTopic:
		return ExchangeTopic
	case interfaces.ExchangeTypeHeaders:
		return ExchangeHeaders
	default:
		return ExchangeDirect
	}
}
