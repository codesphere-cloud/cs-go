// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"fmt"
	"reflect"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

func makeGenericOpenAPIError(body []byte, errStr string) error {
	typ := reflect.TypeOf(openapi_client.GenericOpenAPIError{})
	val := reflect.New(typ).Elem()

	bodyField := val.FieldByName("body")
	errField := val.FieldByName("error")

	reflect.NewAt(bodyField.Type(), unsafe.Pointer(bodyField.UnsafeAddr())).Elem().Set(reflect.ValueOf(body))
	reflect.NewAt(errField.Type(), unsafe.Pointer(errField.UnsafeAddr())).Elem().Set(reflect.ValueOf(errStr))

	return val.Addr().Interface().(error)
}

var _ = Describe("FormatAPIError", func() {
	It("returns nil for nil error", func() {
		Expect(errors.FormatAPIError(nil)).To(BeNil())
	})

	It("returns regular error unchanged", func() {
		err := fmt.Errorf("regular error")
		res := errors.FormatAPIError(err)
		Expect(res).ToNot(BeNil())
		Expect(res.Error()).To(Equal("regular error"))
	})

	It("parses API JSON error and formats it", func() {
		apiErr := makeGenericOpenAPIError([]byte(`{"status":400,"title":"Workspace is not running","detail":"Workspace '796636' is not in a running state.","traceId":"svJDMa5"}`), "400 Bad Request")
		res := errors.FormatAPIError(apiErr)
		Expect(res).ToNot(BeNil())
		Expect(res.Error()).To(Equal("API error 400 Workspace is not running: Workspace '796636' is not in a running state."))
	})
})
