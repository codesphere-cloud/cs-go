// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v2"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("Organization", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		l          cmd.ListOrgCmd
	)

	BeforeEach(func() {
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockClient = cmd.NewMockClient(GinkgoT())
		l = cmd.ListOrgCmd{
			Opts: &cmd.ListOptions{
				GlobalOptions: &cmd.GlobalOptions{
					Env:   mockEnv,
					OrgId: -1, // force using the env mock to get a org ID
				},
			},
			ClientFactory: cmd.NewClient, // Default to real client, will be overridden in specific tests
		}
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("RunE Method", func() {
		It("successful execution", func() {
			mockEnv.EXPECT().GetApiToken().Return("test-token", nil).Maybe()
			mockEnv.EXPECT().GetApiUrl().Return("https://cloud.codesphere.com/api").Maybe()

			// Override the ClientFactory to return our mockClient
			l.ClientFactory = func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return mockClient, nil
			}

			mockClient.EXPECT().ListOrganizations().Return([]api.Organization{}, nil).Once()
			err := l.RunE(nil, []string{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("fails when client creation fails due to missing API token", func() {
			// Mock failure to get API token
			mockEnv.EXPECT().GetApiToken().Return("", errors.New("CS_TOKEN env var required, but not set"))

			err := l.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client"))
		})
		It("fails when client creation fails due to invalid API URL", func() {
			// Mock successful token
			mockEnv.EXPECT().GetApiToken().Return("test-token", nil)
			mockEnv.EXPECT().GetApiUrl().Return("ht tp://invalid url with spaces")

			err := l.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client"))
		})

		It("successfully creates client but fails when the API call fails", func() {
			// Mock successful token and URL to allow NewClient to succeed
			mockEnv.EXPECT().GetApiToken().Return("test-token", nil).Maybe()
			mockEnv.EXPECT().GetApiUrl().Return("https://cloud.codesphere.com/api").Maybe()

			// We should also use the mock factory here to simulate an API failure deterministically
			l.ClientFactory = func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return mockClient, nil
			}
			mockClient.EXPECT().ListOrganizations().Return(nil, errors.New("API error")).Once()

			err := l.RunE(nil, []string{})
			Expect(err).To(MatchError(ContainSubstring("failed to list organizations: API error")))
		})
	})

	Context("ListOrg Method", func() {
		It("successfully lists organizations", func() {
			expectedOrgs := []api.Organization{
				{Id: "d90e5f82-445e-4397-a90e-74d55cd4be3c", Name: "fakeForTeam0"},
				{Id: "d90e5f82-445e-4397-a90e-74d55cd4be4c", Name: "fakeForTeam1"},
			}
			mockClient.EXPECT().ListOrganizations().Return(expectedOrgs, nil).Once()

			orgs, err := l.ListOrganizations(mockClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(Equal(expectedOrgs))
		})

		It("returns an empty list without error when no organizations exist", func() {
			mockClient.EXPECT().ListOrganizations().Return([]api.Organization{}, nil).Once()
			orgs, err := l.ListOrganizations(mockClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(BeEmpty())
		})

		It("successfully lists organizations in JSON format", func() {
			l.Opts.OutputFormat = cmd.OutputFormatJSON
			expectedOrgs := []api.Organization{
				{Id: "d90e5f82-445e-4397-a90e-74d55cd4be3c", Name: "fakeForTeam0"},
				{Id: "d90e5f82-445e-4397-a90e-74d55cd4be4c", Name: "fakeForTeam1"},
			}
			mockClient.EXPECT().ListOrganizations().Return(expectedOrgs, nil).Once()

			// Capture Stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			orgs, err := l.ListOrganizations(mockClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(Equal(expectedOrgs))

			// Restore Stdout
			err = w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(Equal(expectedOrgs))

			// Verify actual JSON output
			var actualOrgs []api.Organization
			err = json.Unmarshal(buf.Bytes(), &actualOrgs)
			Expect(err).NotTo(HaveOccurred(), "The output printed to console was not valid JSON")
			Expect(actualOrgs).To(Equal(expectedOrgs))
		})

		It("successfully lists organizations in YAML format", func() {
			l.Opts.OutputFormat = cmd.OutputFormatYAML
			expectedOrgs := []api.Organization{
				{Id: "d90e5f82-445e-4397-a90e-74d55cd4be3c", Name: "fakeForTeam0"},
				{Id: "d90e5f82-445e-4397-a90e-74d55cd4be4c", Name: "fakeForTeam1"},
			}
			mockClient.EXPECT().ListOrganizations().Return(expectedOrgs, nil).Once()

			// Capture Stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			orgs, err := l.ListOrganizations(mockClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(Equal(expectedOrgs))



			// Restore Stdout
			err = w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(Equal(expectedOrgs))

			// Verify actual YAML output
			var actualOrgs []api.Organization

			err = yaml.Unmarshal(buf.Bytes(), &actualOrgs)
			Expect(err).NotTo(HaveOccurred(), "The output printed to console was not valid YAML")
			Expect(actualOrgs).To(Equal(expectedOrgs))
		})
	})
})

var _ = Describe("AddListOrgCmd", func() {
	var (
		parentCmd *cobra.Command
		listOpts  *cmd.ListOptions
	)

	BeforeEach(func() {
		parentCmd = &cobra.Command{Use: "list"}
		listOpts = &cmd.ListOptions{
			GlobalOptions: &cmd.GlobalOptions{},
		}
	})

	It("adds the org command with correct properties", func() {
		cmd.AddListOrgCmd(parentCmd, listOpts)

		var orgCmd *cobra.Command
		for _, c := range parentCmd.Commands() {
			if c.Use == "org" {
				orgCmd = c
				break
			}
		}

		Expect(orgCmd).NotTo(BeNil())
		Expect(orgCmd.Use).To(Equal("org"))
		Expect(orgCmd.Short).To(Equal("List organizations"))
		Expect(orgCmd.RunE).NotTo(BeNil())
	})
})
