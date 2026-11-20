// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SelectOption prints a numbered list and returns the zero-based index of the
// selected option. An empty list, a blank answer, and EOF cancel the selection.
func SelectOption(input io.Reader, output io.Writer, options []string, prompt string) (index int, selected bool, err error) {
	if len(options) == 0 {
		return 0, false, nil
	}

	var menu strings.Builder
	for i, option := range options {
		fmt.Fprintf(&menu, "%d. %s\n", i+1, option)
	}
	fmt.Fprintf(&menu, "%s, or press Enter to cancel: ", prompt)
	if _, err := io.WriteString(output, menu.String()); err != nil {
		return 0, false, fmt.Errorf("failed to display selection: %w", err)
	}

	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, false, fmt.Errorf("failed to read selection: %w", err)
		}
		return 0, false, nil
	}

	answer := strings.TrimSpace(scanner.Text())
	if answer == "" {
		return 0, false, nil
	}
	number, err := strconv.Atoi(answer)
	if err != nil || number < 1 || number > len(options) {
		return 0, false, fmt.Errorf("invalid selection, enter a number from 1 to %d", len(options))
	}
	return number - 1, true, nil
}
