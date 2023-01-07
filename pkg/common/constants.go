// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package common

import "errors"

/*
	Message Exchange Types
*/
const (
	ExchangeDirect  = "direct"
	ExchangeFanout  = "fanout"
	ExchangeTopic   = "topic"
	ExchangeHeaders = "headers"
)

var (
	// ErrUnresolvedRabbitError unresolvable rabbit error
	ErrUnresolvedRabbitError = errors.New("unresolvable rabbit error")
)
