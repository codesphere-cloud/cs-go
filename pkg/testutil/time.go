// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/mock"
)

func MockTime() *api.MockTime {
	currentTime := time.Unix(1746190963, 0)
	return MockTimeAt(currentTime)
}

func MockTimeAt(t time.Time) *api.MockTime {
	currentTime := t
	m := api.NewMockTime(ginkgo.GinkgoT())
	m.EXPECT().Now().RunAndReturn(func() time.Time {
		return currentTime
	}).Maybe()
	m.EXPECT().Sleep(mock.Anything).Run(func(delay time.Duration) {
		currentTime = currentTime.Add(delay)
	}).Maybe()
	return m
}
