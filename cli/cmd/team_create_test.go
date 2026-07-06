// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateTeam", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.CreateTeamCmd
		teamId     int
		orgId      string
		teamName   string
		dcId       int
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		orgId = "d90e5f82-445e-4397-a90e-74d55cd4be3c"
		teamName = "test-team"
		dcId = 1 // Default data center ID for testing
		c = &cmd.CreateTeamCmd{
			Opts: cmd.CreateTeamOpts{
				GlobalOptions: &cmd.GlobalOptions{
					Env:    mockEnv,
					TeamId: teamId,
					// OrgId is intentionally left empty here, will be set in BeforeEach for specific contexts
				},
				Name: teamName,
				DcId: dcId,
			},
			ClientFactory: func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return mockClient, nil
			},
		}
		// Mock common environment calls needed for client creation
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("Validation", func() {
		It("should fail if the team name is empty", func() {
			team, err := c.CreateTeam(mockClient, orgId, "", dcId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("team name cannot be empty"))
			Expect(team).To(BeNil())
		})
	})

	Context("RunE execution flow", func() {
		It("should successfully create a team when organization ID is provided via environment", func() {
			c.Opts.OrgId = "" // Ensure flag is empty
			mockEnv.EXPECT().GetOrgId().Return(orgId).Once()

			expectedTeam := api.Team{
				Id:                  teamId,
				Name:                teamName,
				OrganizationId:      &orgId,
				DefaultDataCenterId: dcId,
			}
			mockClient.EXPECT().CreateTeam(orgId, teamName, dcId).Return(&expectedTeam, nil).Once()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("Team created: %v in Organization: %v\n", teamId, orgId)))
		})

		It("should successfully create a team and print the correct message when no organization ID is provided", func() {
			c.Opts.OrgId = "" // Ensure flag is empty
			mockEnv.EXPECT().GetOrgId().Return("").Once()

			expectedTeam := api.Team{
				Id:                  teamId,
				Name:                teamName,
				DefaultDataCenterId: dcId,
			}
			mockClient.EXPECT().CreateTeam("", teamName, dcId).Return(&expectedTeam, nil).Once()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("Team created: %v\n", teamId)))
			Expect(buf.String()).ToNot(ContainSubstring("in Organization"))
		})
	})

	Context("when creating a team with an organization ID", func() {
		BeforeEach(func() {
			c.Opts.OrgId = orgId // Set OrgId via GlobalOptions (flag equivalent)
		})

		It("should successfully create the team and return the correct object", func() {
			expectedTeam := api.Team{
				Id:                  teamId,
				Name:                teamName,
				OrganizationId:      &orgId,
				DefaultDataCenterId: dcId,
			}
			// Expect CreateTeam API call with the provided orgId
			mockClient.EXPECT().CreateTeam(orgId, teamName, dcId).Return(&expectedTeam, nil).Once()

			team, err := c.CreateTeam(mockClient, orgId, teamName, dcId)
			Expect(err).ToNot(HaveOccurred())
			Expect(team.Name).To(Equal(teamName))
			Expect(*team.OrganizationId).To(Equal(orgId))
		})
		It("should fail to create with no permission in this orgId", func() {
			// Change the return values to (nil, error) to simulate an API failure
			mockClient.EXPECT().CreateTeam(orgId, teamName, dcId).Return(nil, fmt.Errorf("permission denied")).Once()

			team, err := c.CreateTeam(mockClient, orgId, teamName, dcId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create team: permission denied"))
			Expect(team).To(BeNil())
		})
	})

	Context("when creating a team without an organization ID", func() {
		BeforeEach(func() {
			c.Opts.OrgId = "" // Ensure OrgId is empty from flag
		})

		It("should create the team without orgID ", func() {
			expectedTeam := api.Team{
				Id:                  teamId,
				Name:                teamName,
				DefaultDataCenterId: dcId,
			}
			mockClient.EXPECT().CreateTeam("", teamName, dcId).Return(&expectedTeam, nil).Once()
			mockEnv.EXPECT().GetOrgId().Return("").Once()

			team, err := c.CreateTeam(mockClient, mockEnv.GetOrgId(), teamName, dcId)
			Expect(err).ToNot(HaveOccurred())
			Expect(team.Name).To(Equal(teamName))
		})
	})

	Context("when an invalid organization ID format is provided via flag", func() {
		BeforeEach(func() {
			c.Opts.OrgId = "invalid-uuid-format" // Set an invalid UUID in GlobalOptions
		})

		It("should return an error due to invalid organization ID format", func() {
			// The error should occur before the API call to CreateTeam.
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid organization ID format:"))
		})
	})
})
