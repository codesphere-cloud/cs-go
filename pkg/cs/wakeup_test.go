// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs_test

import (
	"errors"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("WakeUpWorkspace", func() {
	var client *cs.MockClient

	BeforeEach(func() {
		client = cs.NewMockClient(GinkgoT())
	})

	It("returns an already running workspace without scaling", func() {
		workspace := api.Workspace{Id: 42, Name: "running"}
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: true}, nil)

		result, err := cs.WakeUpWorkspace(client, 42, time.Minute)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&workspace))
	})

	It("preserves the configured replica count when waking a workspace", func() {
		workspace := api.Workspace{Id: 42, Name: "stopped", Replicas: 3}
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
		client.EXPECT().ScaleWorkspace(42, 3).Return(nil)
		client.EXPECT().WaitForWorkspaceRunning(mock.Anything, time.Minute).Run(func(value *api.Workspace, _ time.Duration) {
			Expect(value).To(Equal(&workspace))
		}).Return(nil)

		result, err := cs.WakeUpWorkspace(client, 42, time.Minute)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&workspace))
	})

	It("wraps a workspace status error", func() {
		workspace := api.Workspace{Id: 42}
		statusErr := errors.New("status unavailable")
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(nil, statusErr)

		result, err := cs.WakeUpWorkspace(client, 42, time.Minute)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("failed to get workspace status")))
		Expect(errors.Is(err, statusErr)).To(BeTrue())
	})

	It("wraps a wait error", func() {
		workspace := api.Workspace{Id: 42}
		waitErr := errors.New("timed out")
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
		client.EXPECT().ScaleWorkspace(42, 1).Return(nil)
		client.EXPECT().WaitForWorkspaceRunning(mock.Anything, time.Minute).Return(waitErr)

		result, err := cs.WakeUpWorkspace(client, 42, time.Minute)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("workspace did not become running")))
		Expect(errors.Is(err, waitErr)).To(BeTrue())
	})
})
