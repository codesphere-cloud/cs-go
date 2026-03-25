// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

func FormatAPIError(r *http.Response, err error) error {
	if err == nil {
		return nil
	}

	if r == nil {
		r = &http.Response{
			StatusCode: -1,
		}
	}
	if r.Request == nil {
		r.Request = &http.Request{URL: &url.URL{}}
	}

	openAPIErr, ok := err.(*openapi_client.GenericOpenAPIError)
	if !ok {
		return fmt.Errorf("unexpected error %d at URL %s: %w", r.StatusCode, r.Request.URL, err)
	}

	body := openAPIErr.Body()
	if len(body) == 0 {
		return fmt.Errorf("unexpected error %d at URL %s: %w", r.StatusCode, r.Request.URL, err)
	}

	var apiErr APIErrorResponse
	if json.Unmarshal(body, &apiErr) != nil {
		return fmt.Errorf("unexpected error %d at URL %s: %w", r.StatusCode, r.Request.URL, err)
	}

	return fmt.Errorf("codesphere API returned error %d (%s): %s", apiErr.Status, apiErr.Title, apiErr.Detail)
}

// IsRetryable returns true if the error is a server error (HTTP 500, 502, 503, 504).
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, code := range []string{"error 500", "error 502", "error 503", "error 504"} {
		if strings.Contains(msg, code) {
			return true
		}
	}
	return false
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "error 404")
}
