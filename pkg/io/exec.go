package io

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Exec interface {
	ExecuteCommand(ctx context.Context, cmdArgs []string) (int, error)
}

type RealExec struct{}

func (c *RealExec) ExecuteCommand(ctx context.Context, cmdArgs []string) (int, error) {
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
	StreamOutput(&wg, stdoutPipe, os.Stdout)
	StreamOutput(&wg, stderrPipe, os.Stderr)

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

func StreamOutput(wg *sync.WaitGroup, input io.Reader, output io.Writer) {
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
