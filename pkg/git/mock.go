package git

import (
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/mock"
)

// NewMockGit creates a new mock Git service for testing purposes.
func NewMockGit(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGit {
	mock := &MockGit{}
	mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

type MockGit struct {
	mock.Mock
}

func (_m *MockGit) CloneRepository(fs *cs.FileSystem, url string, path string) (*git.Repository, error) {
	ret := _m.Called(fs, url, path)

	var r0 *git.Repository
	var r1 error
	if rf, ok := ret.Get(0).(func(*cs.FileSystem, string, string) (*git.Repository, error)); ok {
		return rf(fs, url, path)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

type MockGit_Expecter struct {
	mock *mock.Mock
}

func (_m *MockGit) EXPECT() *MockGit_Expecter {
	return &MockGit_Expecter{mock: &_m.Mock}
}

type MockGit_CloneRepository_Call struct {
	*mock.Call
}

func (_e *MockGit_Expecter) CloneRepository(fs *cs.FileSystem, url string, path string) *MockGit_CloneRepository_Call {
	return &MockGit_CloneRepository_Call{Call: _e.mock.On("CloneRepository", fs, url, path)}
}
