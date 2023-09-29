package local

import (
	"os"
	"os/exec"
)

type ExecMock struct{}

func (e *ExecMock) Command(name string, arg ...string) *exec.Cmd {
	path, _ := exec.LookPath("echo")
	return &exec.Cmd{
		Path: path,
		Args: []string{},
	}
}

func (e *ExecMock) LookPath(file string) (string, error) {
	return "foo", nil
}

type OSMock struct{}

func (m *OSMock) Chdir(path string) error {
	return nil
}

func (m *OSMock) Stat(name string) (os.FileInfo, error) {
	return os.Stat("/")
}
