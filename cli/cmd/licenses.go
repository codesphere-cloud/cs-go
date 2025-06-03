/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/tmpl"
	"github.com/spf13/cobra"
)

// LicensesCmd represents the licenses command
type LicensesCmd struct {
	cmd *cobra.Command
}

func (c *LicensesCmd) RunE(_ *cobra.Command, args []string) error {
	fmt.Println("Codesphere CLI License:")
	fmt.Println(tmpl.License())

	fmt.Println("=================================")

	fmt.Println("Codesphere CLI licenses included work:")
	fmt.Println(tmpl.Notice())

	return nil
}

func AddLicensesCmd(rootCmd *cobra.Command) {
	licenses := LicensesCmd{
		cmd: &cobra.Command{
			Use:   "licenses",
			Short: "Print license information",
			Long:  `Print information about the Codesphere CLI license and open source licenses of dependencies.`,
		},
	}
	rootCmd.AddCommand(licenses.cmd)
	licenses.cmd.RunE = licenses.RunE
}
