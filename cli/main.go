/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

func main() {
	//Disable printing timestamps on log lines
	log.SetFlags(0)

	cmd.Execute()
}
