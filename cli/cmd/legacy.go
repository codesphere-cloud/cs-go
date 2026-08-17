// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	addcmd "github.com/codesphere-cloud/cs-go/cli/cmd/add"
	createcmd "github.com/codesphere-cloud/cs-go/cli/cmd/create"
	deletecmd "github.com/codesphere-cloud/cs-go/cli/cmd/delete"
	listcmd "github.com/codesphere-cloud/cs-go/cli/cmd/list"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/spf13/cobra"
)

// legacy.go keeps pre-"verb folder" command paths working as deprecated
// aliases, so existing scripts and muscle memory don't break outright.
//
// Cobra's Command.Aliases only covers alternate names at the *same* tree
// position. The restructuring moved several commands to a different
// parent (e.g. `cs team create` -> `cs create team`), which Aliases can't
// express - the old path needs a second, real *cobra.Command.
//
// To do that without duplicating any flag wiring or business logic, each
// legacy leaf is built by calling the real AddXxxCmd constructor against a
// throwaway parent, lifted out of it, and re-parented under its old path
// with its old name and a Deprecated notice. Cobra hides Deprecated
// commands from help/usage output but still executes them normally, and
// excludes them from generated docs (see hack/gendocs).
//
// Remove this file once users have had a chance to migrate.

// relocate builds a command via build against a scratch parent, detaches
// it, and returns it renamed to oldUse with a deprecation notice pointing
// at newPath.
func relocate(oldUse, newPath string, build func(scratch *cobra.Command)) *cobra.Command {
	scratch := &cobra.Command{}
	build(scratch)

	real := scratch.Commands()[0]
	scratch.RemoveCommand(real)

	real.Use = oldUse
	real.Deprecated = fmt.Sprintf("moved, use '%s' instead", newPath)
	return real
}

// AddLegacyCmds re-adds pre-refactor command paths as deprecated aliases
// for the commands that moved to a different parent.
func AddLegacyCmds(rootCmd *cobra.Command, opts shared.RootOptions) {
	// `list team-members` and `log` used to take a local --output flag
	// rather than inheriting one from the `list` parent; give the legacy
	// commands their own so `-o json`/`-o yaml` keep working too.
	listOpts := &listcmd.ListOptions{RootOptions: opts}

	team := &cobra.Command{
		Use:        "team",
		Short:      "Manage Team",
		Deprecated: "moved, use 'create team' / 'delete team' / 'add team-member' / 'delete team-member' / 'list team-members' instead",
	}
	shared.AddCmd(rootCmd, team)
	shared.AddCmd(team, relocate("create", "create team", func(s *cobra.Command) { createcmd.AddCreateTeamCmd(s, opts) }))
	shared.AddCmd(team, relocate("remove", "delete team", func(s *cobra.Command) { deletecmd.AddDeleteTeamCmd(s, opts) }))

	member := &cobra.Command{
		Use:        "member",
		Short:      "Manage team members",
		Deprecated: "moved, use 'add team-member' / 'delete team-member' / 'list team-members' instead",
	}
	shared.AddCmd(team, member)
	shared.AddCmd(member, relocate("add", "add team-member", func(s *cobra.Command) { addcmd.AddAddTeamMemberCmd(s, opts) }))
	shared.AddCmd(member, relocate("remove", "delete team-member", func(s *cobra.Command) { deletecmd.AddDeleteTeamMemberCmd(s, opts) }))

	list := relocate("list", "list team-members", func(s *cobra.Command) { listcmd.AddListTeamMembersCmd(s, listOpts) })
	list.Flags().StringVarP((*string)(&listOpts.OutputFormat), "output", "o", "table", "Output format (table, json, yaml)")
	shared.AddCmd(member, list)

	setEnv := relocate("set-env", "create env", func(s *cobra.Command) { createcmd.AddCreateEnvCmd(s, opts) })
	shared.AddCmd(rootCmd, setEnv)

	log := relocate("log", "list landscape-logs", func(s *cobra.Command) { listcmd.AddListLandscapeLogsCmd(s, listOpts) })
	shared.AddCmd(rootCmd, log)
}
