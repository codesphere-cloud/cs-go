// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
)

func GetTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	return t
}

// This variable is injected during docs generation. Update in Makefile when moving
var binName string

func BinName() string {
	if binName == "" {
		binName = os.Args[0]
	}
	return binName
}

type Example struct {
	Cmd  string
	Desc string
}

func FormatExampleCommands(command string, examples []Example) (res string) {
	for i, ex := range examples {
		if i > 0 {
			res += "\n\n"
		}
		res += fmt.Sprintf("# %s\n$ %s %s %s", ex.Desc, BinName(), command, ex.Cmd)
	}
	return
}

// Remove tabs to allow formatted multi-line descriptions in Code without cluttering
// the help output
func Long(in string) string {
	re := regexp.MustCompile("\n\t+")
	return re.ReplaceAllString(in, "\n")
}

type Prompt struct{}

// Prompt for non-empty user input from STDIN
func (p *Prompt) InputPrompt(prompt string) string {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s: ", prompt)
		input, err := r.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" || err != nil {
			return input
		}
	}
}

func ExecuteCommand(ctx context.Context, cmdArgs []string) (int, error) {
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("error creating stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return -1, fmt.Errorf("error creating stderr pipe: %w", err)
	}

	// Create WaitGroup to ensure all goroutines finish before returning
	var wg sync.WaitGroup
	streamOutput(&wg, stdoutPipe, os.Stdout)
	streamOutput(&wg, stderrPipe, os.Stderr)

	err = cmd.Start()
	if err != nil {
		return -1, fmt.Errorf("error starting command %v: %w", cmdArgs, err)
	}

	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return -1, fmt.Errorf("command %v failed with error: %w", cmdArgs, err)
	}

	wg.Wait()
	return 0, nil
}

func streamOutput(wg *sync.WaitGroup, input io.ReadCloser, output *os.File) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			_, err := fmt.Fprintln(output, scanner.Text())
			if err != nil {
				fmt.Printf("error reading input: %v\n", err)
			}
		}
		err := scanner.Err()
		if err != nil && err != io.EOF { // io.EOF is expected at end of stream
			fmt.Printf("error reading input: %v\n", err)
		}
	}()
}
