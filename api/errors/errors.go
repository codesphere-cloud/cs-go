// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"time"
)

type TimedOutError struct {
	msg string
}

func (e *TimedOutError) Error() string {
	return e.msg
}

func TimedOut(operation string, timeout time.Duration) *TimedOutError {
	return &TimedOutError{
		msg: fmt.Sprintf("%s timed out after %s", operation, timeout.String()),
	}
}

type NotFoundError struct {
	msg string
}

func (e *NotFoundError) Error() string {
	return e.msg
}

func NotFound(msg string) *NotFoundError {
	return &NotFoundError{
		msg: msg,
	}
}

type DuplicatedError struct {
	msg string
}

func (e *DuplicatedError) Error() string {
	return e.msg
}

func Duplicated(msg string) *DuplicatedError {
	return &DuplicatedError{
		msg: msg,
	}
}
