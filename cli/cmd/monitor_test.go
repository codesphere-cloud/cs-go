// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
)

var _ = Describe("Monitor", func() {
	var (
		mockHttpServer *csio.MockHttpServer
		mockTime       *api.MockTime
		mockExec       *csio.MockExec
		c              *cmd.MonitorCmd
		listenAddress  string
		maxRestarts    int
		currentTime    time.Time
		forward        string
		skipTLS        bool
		caCertFile     string
	)

	BeforeEach(func() {
		mockHttpServer = csio.NewMockHttpServer(GinkgoT())
		mockTime = api.NewMockTime(GinkgoT())
		mockExec = csio.NewMockExec(GinkgoT())
		maxRestarts = 0 //to make tests finite
		listenAddress = ":3000"
		skipTLS = true

	})
	JustBeforeEach(func() {
		c = &cmd.MonitorCmd{
			Time: mockTime,
			Http: mockHttpServer,
			Exec: mockExec,
			Opts: cmd.MonitorOpts{
				ListenAddress:      &listenAddress,
				MaxRestarts:        &maxRestarts,
				Forward:            &forward,
				InsecureSkipVerify: &skipTLS,
				CaCertFile:         &caCertFile,
			},
		}

		currentTime = time.Unix(1746190963, 0)
		mockTime.EXPECT().Now().RunAndReturn(func() time.Time {
			return currentTime
		}).Maybe()
		mockTime.EXPECT().Sleep(mock.Anything).Run(func(t time.Duration) {
			currentTime = currentTime.Add(t)
		}).Maybe()

		mockHttpServer.EXPECT().ListenAndServe(mock.Anything, mock.Anything).Return(nil).Maybe()
	})

	Context("With default healthcheck endpoint", func() {
		JustBeforeEach(func() {
			mockHttpServer.EXPECT().Handle(mock.Anything, mock.Anything)
			mockHttpServer.EXPECT().HandleFunc(mock.Anything, mock.Anything)
		})
		Context("Command exits after 10 seconds with exit code 0", func() {
			It("Doesn't return an error", func() {
				mockExec.EXPECT().ExecuteCommand(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, args []string) (int, error) {
						mockTime.Sleep(10 * time.Second)
						return 0, nil
					})

				err := c.RunCommandWithHealthcheck(context.TODO(), []string{"fake-sleep", "10s"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("Command exits after 0.3 seconds", func() {
			BeforeEach(func() {
				maxRestarts = 1
			})
			Context("Command doesn't return an error", func() {
				It("Restarts immediately", func() {
					mockExec.EXPECT().ExecuteCommand(mock.Anything, mock.Anything).RunAndReturn(
						func(ctx context.Context, args []string) (int, error) {
							log.Println("lol")
							mockTime.Sleep(300 * time.Millisecond)
							return 0, nil
						}).Twice()

					err := c.RunCommandWithHealthcheck(context.TODO(), []string{"fake-sleep", "300ms"})
					Expect(err).NotTo(HaveOccurred())

				})

			})
			Context("Command immediately returns an error", func() {
				It("Restarts after a delay of 5 seconds", func() {

					mockExec.EXPECT().ExecuteCommand(mock.Anything, mock.Anything).RunAndReturn(
						func(ctx context.Context, args []string) (int, error) {
							return 1, nil
						}).Twice()

					startTime := currentTime
					err := c.RunCommandWithHealthcheck(context.TODO(), []string{"fake-exit", "1"})
					endTime := currentTime

					Expect(err).NotTo(HaveOccurred())
					Expect(endTime.Sub(startTime)).To(Equal(5 * time.Second))
				})
			})
		})
	})

	Context("With healthcheck forwarding", func() {
		BeforeEach(func() {
			forward = "http://localhost:3000"
		})
		Context("Command exits after 10 seconds with exit code 0", func() {
			It("Doesn't return an error", func() {
				mockExec.EXPECT().ExecuteCommand(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, args []string) (int, error) {
						mockTime.Sleep(10 * time.Second)
						return 0, nil
					})

				err := c.RunCommandWithHealthcheck(context.TODO(), []string{"fake-sleep", "10s"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

})
