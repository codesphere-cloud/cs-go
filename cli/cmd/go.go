// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"io"
	"log"
	"strings"

	"github.com/codesphere-cloud/cs-go/pkg/tmpl"
	"github.com/spf13/cobra"
)

type GoCmd struct {
	cmd *cobra.Command
}

func (c *GoCmd) RunE(_ *cobra.Command, args []string) error {
	x, _ := gzip.NewReader(strings.NewReader(tmpl.GoData()))
	s, _ := io.ReadAll(x)
	d, _ := base64.StdEncoding.DecodeString(string(s))
	log.Print(string(d))
	return nil
}

func AddGoCmd(rootCmd *cobra.Command) {
	goCmd := GoCmd{
		cmd: &cobra.Command{Hidden: true, Use: "go"},
	}
	rootCmd.AddCommand(goCmd.cmd)
	goCmd.cmd.RunE = goCmd.RunE
}
