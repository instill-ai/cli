package local

import (
	"os"
	"os/exec"
)

type ExecMock struct{}

func (e *ExecMock) Command(name string, arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: "/usr/sbin/ls",
		Args: []string{"/"},
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
