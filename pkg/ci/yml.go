package ci

import (
	"fmt"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	"gopkg.in/yaml.v2"
)

type CiYml struct {
	Prepare Steps              `yaml:"prepare"`
	Test    Steps              `yaml:"test"`
	Run     map[string]Service `yaml:"run"`
}

type Steps struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type Service struct {
	Steps    []Step  `yaml:"steps"`
	Plan     int     `yaml:"plan"`
	Replicas int     `yaml:"replicas"`
	IsPublic bool    `yaml:"isPublic"`
	Network  Network `yaml:"network"`
}

type Network struct {
	Path      string `yaml:"path"`
	StripPath bool   `yaml:"stripPath"`
	Paths     []Path `yaml:"paths"`
	Ports     []Port `yaml:"ports"`
}

type Path struct {
	Port      int    `yaml:"port"`
	Path      string `yaml:"path"`
	StripPath bool   `yaml:"stripPath"`
}

type Port struct {
	Port     int  `yaml:"port"`
	IsPublic bool `yaml:"isPublic"`
}

func ReadYmlFile(fs billy.Filesystem, path string) (*CiYml, error) {
	ymlFile, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error reading yml file: %w", err)
	}

	ymlFileBytes, err := util.ReadFile(fs, ymlFile.Name())
	if err != nil {
		return nil, fmt.Errorf("error reading yml file: %w", err)
	}

	err = ymlFile.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing yml file: %w", err)
	}

	ymlContent := &CiYml{}
	err = yaml.Unmarshal(ymlFileBytes, &ymlContent)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yml file: %w", err)
	}

	return ymlContent, nil
}
