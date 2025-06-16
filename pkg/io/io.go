// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

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

type HttpServer interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
	Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
	ListenAndServe(addr string, handler http.Handler) error
}

type RealHttpServer struct{}

func (*RealHttpServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (*RealHttpServer) Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}
func (*RealHttpServer) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

func (*RealHttpServer) ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}
