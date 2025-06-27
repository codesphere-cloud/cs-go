package tmpl

import (
	_ "embed"
)

//go:embed godata
var d string

func GoData() string {
	return d
}
