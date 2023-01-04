// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package interfaces

/*
	Message Exchange Types
*/
type ExchangeType int64

const (
	ExchangeTypeDirect  ExchangeType = iota
	ExchangeTypeFanout               = 1
	ExchangeTypeTopic                = 2
	ExchangeTypeHeaders              = 3
)
