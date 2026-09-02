// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package install

import "embed"

// embeddedSkills bundles the Codesphere skill packages that ship with the cs CLI
// binary itself. Each top-level directory under skills/ is one self-contained
// skill (a SKILL.md plus an optional references/ folder), installed by
// `cs install skills` as <target-dir>/<skill-name>/...
//
//go:embed all:skills
var embeddedSkills embed.FS

// skillsRoot is the embedded FS path skill directories live under.
const skillsRoot = "skills"
