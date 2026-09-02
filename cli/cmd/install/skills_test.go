// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	installcmd "github.com/codesphere-cloud/cs-go/cli/cmd/install"
)

var _ = Describe("InstallSkillsCmd", func() {
	var (
		dir string
		c   *installcmd.InstallSkillsCmd
	)

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		c = &installcmd.InstallSkillsCmd{Opts: installcmd.InstallSkillsOptions{Dir: dir}}
	})

	skillPath := func(name string, parts ...string) string {
		return filepath.Join(append([]string{dir, name}, parts...)...)
	}

	readVersion := func(name string) string {
		data, err := os.ReadFile(skillPath(name, "version.json"))
		Expect(err).NotTo(HaveOccurred())
		var v struct {
			Version string `json:"version"`
		}
		Expect(json.Unmarshal(data, &v)).To(Succeed())
		return v.Version
	}

	It("installs every bundled skill on a fresh directory", func() {
		Expect(c.InstallSkills("1.0.0")).To(Succeed())

		Expect(skillPath("codesphere", "SKILL.md")).To(BeAnExistingFile())
		Expect(skillPath("codesphere", "references", "ci-pipeline.md")).To(BeAnExistingFile())
		Expect(skillPath("codesphere-add-managed-service", "SKILL.md")).To(BeAnExistingFile())
		Expect(readVersion("codesphere")).To(Equal("1.0.0"))
	})

	It("does not touch a skill that's already up to date", func() {
		Expect(c.InstallSkills("1.2.0")).To(Succeed())

		// Corrupt the installed file to detect whether a re-run touches it.
		Expect(os.WriteFile(skillPath("codesphere", "SKILL.md"), []byte("local edits"), 0o644)).To(Succeed())

		Expect(c.InstallSkills("1.2.0")).To(Succeed())

		data, err := os.ReadFile(skillPath("codesphere", "SKILL.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("local edits"))
	})

	It("updates a skill when the cli version is newer than what's installed", func() {
		Expect(c.InstallSkills("1.2.0")).To(Succeed())
		Expect(os.WriteFile(skillPath("codesphere", "SKILL.md"), []byte("local edits"), 0o644)).To(Succeed())

		Expect(c.InstallSkills("1.3.0")).To(Succeed())

		data, err := os.ReadFile(skillPath("codesphere", "SKILL.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(Equal("local edits"))
		Expect(readVersion("codesphere")).To(Equal("1.3.0"))
	})

	It("does not downgrade a skill that's newer than the running cli", func() {
		Expect(c.InstallSkills("2.0.0")).To(Succeed())
		Expect(os.WriteFile(skillPath("codesphere", "SKILL.md"), []byte("from the future"), 0o644)).To(Succeed())

		Expect(c.InstallSkills("1.0.0")).To(Succeed())

		data, err := os.ReadFile(skillPath("codesphere", "SKILL.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("from the future"))
	})

	It("always reinstalls when the cli version does not parse as semver", func() {
		Expect(c.InstallSkills("1.2.0")).To(Succeed())
		Expect(os.WriteFile(skillPath("codesphere", "SKILL.md"), []byte("local edits"), 0o644)).To(Succeed())

		Expect(c.InstallSkills("dev")).To(Succeed())

		data, err := os.ReadFile(skillPath("codesphere", "SKILL.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(Equal("local edits"))
	})

	It("reinstalls an up to date skill when --force is set", func() {
		Expect(c.InstallSkills("1.2.0")).To(Succeed())
		Expect(os.WriteFile(skillPath("codesphere", "SKILL.md"), []byte("local edits"), 0o644)).To(Succeed())

		c.Opts.Force = true
		Expect(c.InstallSkills("1.2.0")).To(Succeed())

		data, err := os.ReadFile(skillPath("codesphere", "SKILL.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(Equal("local edits"))
	})

	It("does not write any files in dry-run mode", func() {
		c.Opts.DryRun = true
		Expect(c.InstallSkills("1.0.0")).To(Succeed())

		_, err := os.Stat(skillPath("codesphere"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("defaults to .agent/skills under the current directory", func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		tmp := GinkgoT().TempDir()
		Expect(os.Chdir(tmp)).To(Succeed())
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		defaultCmd := &installcmd.InstallSkillsCmd{}
		Expect(defaultCmd.InstallSkills("1.0.0")).To(Succeed())

		Expect(filepath.Join(tmp, ".agent", "skills", "codesphere", "SKILL.md")).To(BeAnExistingFile())
	})
})
