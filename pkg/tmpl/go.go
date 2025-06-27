// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package tmpl

import (
	_ "embed"
)

//go:embed godata
var d string

func GoData() string {
	return d
}
