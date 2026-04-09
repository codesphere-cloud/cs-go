// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
)

//go:embed docker_ubuntu.tmpl
var dockerTemplateFile string

//go:embed docker_fedora.tmpl
var dockerFedoraTemplateFile string

//go:embed docker_alpine.tmpl
var dockerAlpineTemplateFile string

type DockerTemplateConfig struct {
	BaseImage    string
	PrepareSteps []ci.Step
	Entrypoint   string
}

func CreateDockerfile(config DockerTemplateConfig) ([]byte, error) {
	if len(strings.TrimSpace(config.BaseImage)) == 0 {
		return nil, fmt.Errorf("base image is required")
	}

	templ := dockerTemplateFile
	alpineRe := regexp.MustCompile(".*alpine.*")
	fedoraRe := regexp.MustCompile(".*(fedora)|(coreos)|(rhel).*")
	if alpineRe.MatchString(config.BaseImage) {
		templ = dockerAlpineTemplateFile
		log.Println("Alpine found in " + config.BaseImage)
	}
	if fedoraRe.MatchString(config.BaseImage) {
		templ = dockerFedoraTemplateFile
		log.Println("Fedora found in " + config.BaseImage)
	}
	dockerTemplate, err := template.New("dockerTemplate").Parse(templ)
	if err != nil {
		return nil, fmt.Errorf("error parsing docker template: %w", err)
	}

	var buf bytes.Buffer
	err = dockerTemplate.Execute(&buf, config)
	if err != nil {
		return nil, fmt.Errorf("error executing docker template: %w", err)
	}

	return buf.Bytes(), nil
}
