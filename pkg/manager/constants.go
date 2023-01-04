// Copyright 2023. All Rights Reserved.
// SPDX-License-Identifier: MIT

package manager

import (
	"errors"
)

var (
	// ErrInvalidInput required input was not found
	ErrInvalidInput = errors.New("required input was not found")

	// ErrPublisherNotFound the rabbit publisher was not found
	ErrPublisherNotFound = errors.New("the rabbit publisher was not found")

	// ErrSubscriberNotFound the rabbit publisher was not found
	ErrSubscriberNotFound = errors.New("the rabbit subscriber was not found")
)
