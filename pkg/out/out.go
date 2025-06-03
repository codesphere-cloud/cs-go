// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package out

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

func GetTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	return t
}

var binName string

func FormatExampleCommands(command string, examples map[string]string) (res string) {
	if binName == "" {
		binName = os.Args[0]
	}
	first := true
	for subcommand, comment := range examples {
		if !first {
			res += "\n"
		}
		res += fmt.Sprintf("# %s\n$ %s %s %s\n", comment, binName, command, subcommand)
		first = false
	}
	return
}
