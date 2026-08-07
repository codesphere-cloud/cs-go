// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add

type Client interface {
	AddTeamMember(teamId int, email string, role int) error
}
