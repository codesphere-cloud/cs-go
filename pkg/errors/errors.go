package errors

import (
	"fmt"
	"time"
)

type TimedOut struct {
	msg string
}

func (e *TimedOut) Error() string {
	return e.msg
}

func NewTimedOut(operation string, timeout time.Duration) *TimedOut {
	return &TimedOut{
		msg: fmt.Sprintf("%s timed out after %s", operation, timeout.String()),
	}
}

type NotFound struct {
	msg string
}

func (e *NotFound) Error() string {
	return e.msg
}

func NewNotFound(msg string) *NotFound {
	return &NotFound{
		msg: msg,
	}
}

type Duplicated struct {
	msg string
}

func (e *Duplicated) Error() string {
	return e.msg
}

func NewDuplicated(msg string) *Duplicated {
	return &Duplicated{
		msg: msg,
	}
}
