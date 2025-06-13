package export

import (
	"github.com/stretchr/testify/mock"
)

// NewMockExporter creates a new instance of MockExporter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockExporter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockExporter {
	mock := &MockExporter{}
	mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

type MockExporter struct {
	mock.Mock
}

func (_m *MockExporter) ExportDockerArtifacts(inputPath string, outputPath string, baseImage string, envVars []string) error {
	ret := _m.Called(inputPath, outputPath, baseImage, envVars)

	if len(ret) == 0 {
		panic("no return value specified for ExportDockerArtifacts")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, []string) error); ok {
		r0 = rf(inputPath, outputPath, baseImage, envVars)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type MockExporter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockExporter) EXPECT() *MockExporter_Expecter {
	return &MockExporter_Expecter{mock: &_m.Mock}
}

type MockExporter_ExportDockerArtifacts_Call struct {
	*mock.Call
}

func (_e *MockExporter_Expecter) ExportDockerArtifacts(inputPath string, outputPath string, baseImage string, envVars []string) *MockExporter_ExportDockerArtifacts_Call {
	return &MockExporter_ExportDockerArtifacts_Call{Call: _e.mock.On("ExportDockerArtifacts", inputPath, outputPath, baseImage, envVars)}
}
