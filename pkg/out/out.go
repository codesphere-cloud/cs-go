// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package out

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

func GetTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	return t
}
