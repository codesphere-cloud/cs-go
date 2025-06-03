package main

import (
	"log"

	csgo "github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(csgo.GetRootCmd(), "docs")
	if err != nil {
		log.Fatal(err)
	}
}
