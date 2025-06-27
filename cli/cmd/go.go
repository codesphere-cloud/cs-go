package cmd

import (
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
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
	fmt.Print(string(d))
	return nil
}

func AddGoCmd(rootCmd *cobra.Command) {
	goCmd := GoCmd{
		cmd: &cobra.Command{Hidden: true, Use: "go"},
	}
	rootCmd.AddCommand(goCmd.cmd)
	goCmd.cmd.RunE = goCmd.RunE
}
