package cs

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/util"
	"go.yaml.in/yaml/v2"
)

type DomainType string

const (
	PrivateDevDomain DomainType = "private"
	PublicDevDomain  DomainType = "public"
)

type RepoAccess string

const (
	PublicRepo  RepoAccess = "public"
	PrivateRepo RepoAccess = "private"
)

// UpState represents the state of the up command that is stored in .cs-up.yaml to be reused for subsequent runs.
type UpState struct {
	Profile       string
	Timeout       time.Duration
	Branch        string     `yaml:"branch"`
	TeamId        int        `yaml:"team"`
	WorkspaceId   int        `yaml:"workspace"`
	Plan          int        `yaml:"plan"`
	WorkspaceName string     `yaml:"workspace_name"`
	BaseImage     string     `yaml:"base_image"`
	Env           []string   `yaml:"env"`
	DomainType    DomainType `yaml:"public_dev_domain"`
	RepoAccess    RepoAccess `yaml:"repo_access"`
	Remote        string     `yaml:"remote"`

	StateFile string `yaml:"-"`
	fs        *util.FileSystem
}

// Writes .cs-up.yaml
func (s *UpState) Save() error {
	log.Printf("Saving state to %s", s.StateFile)
	output, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	err = s.fs.WriteFile("", s.StateFile, output, true)
	if err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}

// Load loads the state from local file. If the file doesn't exist, it returns an empty state.
// LoadState only sets properties that are not already set to allow overriding with flags.
// For example, if the user specifies a profile with the --profile flag, it will not be overridden by the value in the state file.
// If the state is updated by the flags provided, the updated state will be saved to the file after the deployment is successful.
func (s *UpState) Load(filename string, t api.Time, fs *util.FileSystem) error {
	s.StateFile = filename
	s.fs = fs

	branchName := fmt.Sprintf("cs-up-%s", t.Now().Format("20060102150405"))
	newState := &UpState{
		DomainType:    PublicDevDomain,
		RepoAccess:    PublicRepo,
		WorkspaceName: branchName,
		Remote:        "origin",
		Profile:       "ci.yml",
		Branch:        branchName,
		Timeout:       5 * time.Minute,
		TeamId:        -1,
		WorkspaceId:   -1,
		Plan:          -1,
	}
	_, err := fs.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat state file: %w", err)
	}
	if err == nil {
		data, err := fs.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read state file: %w", err)
		}
		err = yaml.Unmarshal(data, newState)
		if err != nil {
			return fmt.Errorf("failed to unmarshal state: %w", err)
		}
	}

	if s.Profile == "" {
		s.Profile = newState.Profile
	}
	if s.Timeout == 0 {
		s.Timeout = newState.Timeout
	}
	if s.Branch == "" {
		s.Branch = newState.Branch
	}
	if s.WorkspaceId <= 0 {
		s.WorkspaceId = newState.WorkspaceId
	}
	if s.Plan <= 0 {
		s.Plan = newState.Plan
	}
	if s.WorkspaceName == "" {
		s.WorkspaceName = newState.WorkspaceName
	}
	if len(s.Env) == 0 {
		s.Env = newState.Env
	}
	if s.BaseImage == "" {
		s.BaseImage = newState.BaseImage
	}
	if s.DomainType == "" {
		s.DomainType = newState.DomainType
	}
	if s.RepoAccess == "" {
		s.RepoAccess = newState.RepoAccess
	}
	if s.Remote == "" {
		s.Remote = newState.Remote
	}
	if s.TeamId <= 0 {
		s.TeamId = newState.TeamId
	}
	return nil
}
