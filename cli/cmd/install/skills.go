// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

// versionFileName is written into every installed skill folder, recording the
// cs CLI version that last installed or updated it.
const versionFileName = "version.json"

// skillVersion is the schema of a skill folder's version.json.
type skillVersion struct {
	Version string `json:"version"`
}

// InstallSkillsOptions holds the flags of `cs install skills`. Exported (rather
// than plain unexported fields) so tests can drive InstallSkillsCmd directly
// without going through cobra flag parsing.
type InstallSkillsOptions struct {
	// Dir is the target skills directory. Empty means the default: ".agent/skills"
	// in the current directory, or in the user's home directory if Global is set.
	Dir    string
	Global bool
	Force  bool
	DryRun bool
}

type InstallSkillsCmd struct {
	Opts InstallSkillsOptions
	cmd  *cobra.Command
}

func AddInstallSkillsCmd(parent *cobra.Command) {
	c := InstallSkillsCmd{
		cmd: &cobra.Command{
			Use:   "skills",
			Short: "Install the Codesphere skills bundled with this CLI",
			Long: csio.Long(`Installs the Codesphere skill package(s) bundled with this cs CLI release
				into a local skills folder (".agent/skills" by default), for use by AI coding
				agents that support the Agent Skills format.

				Every installed skill folder gets a version.json recording which cs CLI version
				last installed or updated it. Re-running this command compares that installed
				version against the running cs CLI's own version and only overwrites a skill
				when the CLI is newer - so it's safe to run again after "cs update" to pick up
				newer skills, and a no-op otherwise.`),
			Example: csio.FormatExampleCommands("install skills", []csio.Example{
				{Cmd: "", Desc: "Install/update all bundled skills into ./.agent/skills"},
				{Cmd: "--global", Desc: "Install/update into ~/.agent/skills instead"},
				{Cmd: "--dir path/to/skills", Desc: "Install/update into a custom directory"},
				{Cmd: "--force", Desc: "Reinstall even if already up to date"},
			}),
		},
	}
	c.cmd.Flags().StringVar(&c.Opts.Dir, "dir", "", `Target skills directory (defaults to ".agent/skills", or "~/.agent/skills" with --global)`)
	c.cmd.Flags().BoolVar(&c.Opts.Global, "global", false, "Install into the user's home directory instead of the current directory")
	c.cmd.Flags().BoolVar(&c.Opts.Force, "force", false, "Reinstall every skill even if it's already up to date")
	c.cmd.Flags().BoolVar(&c.Opts.DryRun, "dry-run", false, "Show what would be installed without writing any files")
	c.cmd.RunE = c.RunE
	shared.AddCmd(parent, c.cmd)
}

func (c *InstallSkillsCmd) RunE(_ *cobra.Command, _ []string) error {
	return c.InstallSkills(cs.Version())
}

// InstallSkills installs or updates every skill bundled with this binary into
// the configured target directory, treating cliVersion as "the version of this
// cs CLI build" for the up-to-date comparison against each skill's installed
// version.json. It takes cliVersion explicitly (rather than always calling
// cs.Version() itself) so it's deterministically testable.
func (c *InstallSkillsCmd) InstallSkills(cliVersion string) error {
	target, err := c.resolveTargetDir()
	if err != nil {
		return err
	}

	names, err := listEmbeddedSkills()
	if err != nil {
		return err
	}

	for _, name := range names {
		if err := installSkill(embeddedSkills, name, target, cliVersion, c.Opts.Force, c.Opts.DryRun); err != nil {
			return fmt.Errorf("installing skill %q: %w", name, err)
		}
	}
	return nil
}

func (c *InstallSkillsCmd) resolveTargetDir() (string, error) {
	if c.Opts.Dir != "" {
		return c.Opts.Dir, nil
	}
	if c.Opts.Global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".agent", "skills"), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolving current directory: %w", err)
	}
	return filepath.Join(cwd, ".agent", "skills"), nil
}

// listEmbeddedSkills returns the names of every skill bundled with this binary, sorted.
func listEmbeddedSkills() ([]string, error) {
	entries, err := embeddedSkills.ReadDir(skillsRoot)
	if err != nil {
		return nil, fmt.Errorf("reading bundled skills: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// installSkill installs or updates a single bundled skill into targetBase/<name>,
// gated on whether cliVersion is newer than whatever's already installed there.
func installSkill(bundled fs.FS, name, targetBase, cliVersion string, force, dryRun bool) error {
	destDir := filepath.Join(targetBase, name)

	installed, wasInstalled := readInstalledVersion(destDir)
	needsUpdate := force || !wasInstalled || isNewer(cliVersion, installed)

	if !needsUpdate {
		log.Printf("%s: already up to date (%s)\n", name, installed)
		return nil
	}

	if dryRun {
		if wasInstalled {
			log.Printf("%s: would update %s -> %s\n", name, installed, cliVersion)
		} else {
			log.Printf("%s: would install (%s)\n", name, cliVersion)
		}
		return nil
	}

	srcDir := skillsRoot + "/" + name
	if err := copyFS(bundled, srcDir, destDir); err != nil {
		return err
	}
	if err := writeInstalledVersion(destDir, cliVersion); err != nil {
		return err
	}

	if wasInstalled {
		log.Printf("%s: updated %s -> %s\n", name, installed, cliVersion)
	} else {
		log.Printf("%s: installed (%s)\n", name, cliVersion)
	}
	return nil
}

// readInstalledVersion reads the version.json of an already-installed skill folder.
// The second return value is false when the skill isn't installed yet, or its
// version.json is missing/unreadable.
func readInstalledVersion(destDir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(destDir, versionFileName))
	if err != nil {
		return "", false
	}
	var v skillVersion
	if err := json.Unmarshal(data, &v); err != nil || v.Version == "" {
		return "", false
	}
	return v.Version, true
}

func writeInstalledVersion(destDir, version string) error {
	data, err := json.MarshalIndent(skillVersion{Version: version}, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", versionFileName, err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(destDir, versionFileName), data, 0o644)
}

// isNewer reports whether cliVersion is a strictly newer semver than installed.
// Either version failing to parse (e.g. a local "dev" build, or a hand-edited
// version.json) is treated as "needs an update" rather than blocking one.
func isNewer(cliVersion, installed string) bool {
	current, err := semver.NewVersion(cliVersion)
	if err != nil {
		return true
	}
	previous, err := semver.NewVersion(installed)
	if err != nil {
		return true
	}
	return current.GreaterThan(previous)
}

// copyFS recursively copies the srcDir subtree of an fs.FS onto destDir on disk.
func copyFS(bundled fs.FS, srcDir, destDir string) error {
	return fs.WalkDir(bundled, srcDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel := strings.TrimPrefix(strings.TrimPrefix(p, srcDir), "/")
		target := filepath.Join(destDir, filepath.FromSlash(rel))

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := fs.ReadFile(bundled, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
