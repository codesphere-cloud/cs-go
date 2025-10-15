// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/api/openapi_client"
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

type APIErrorResponse struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	TraceId string `json:"traceId"`
}

func FormatAPIError(err error) error {
	if err == nil {
		return nil
	}

	openAPIErr, ok := err.(*openapi_client.GenericOpenAPIError)
	if !ok {
		return err
	}

	body := openAPIErr.Body()
	if len(body) == 0 {
		return err
	}

	var apiErr APIErrorResponse
	if json.Unmarshal(body, &apiErr) != nil {
		return err
	}

	return fmt.Errorf("codesphere API returned error %d (%s): %s", apiErr.Status, apiErr.Title, apiErr.Detail)
}
