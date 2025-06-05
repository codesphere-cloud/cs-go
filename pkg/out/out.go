// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package out

import (
	"fmt"
	"os"
	"regexp"

	"github.com/jedib0t/go-pretty/v6/table"
)

func GetTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	return t
}

var binName string

type Example struct {
	Cmd  string
	Desc string
}

func FormatExampleCommands(command string, examples []Example) (res string) {
	if binName == "" {
		binName = os.Args[0]
	}
	for i, ex := range examples {
		if i > 0 {
			res += "\n\n"
		}
		res += fmt.Sprintf("# %s\n$ %s %s %s", ex.Desc, binName, command, ex.Cmd)
	}
	return
}

// Remove tabs to allow formatted multi-line descriptions in Code without cluttering
// the help output
func Long(in string) string {
	re := regexp.MustCompile("\n\t+")
	return re.ReplaceAllString(in, "\n")
}
