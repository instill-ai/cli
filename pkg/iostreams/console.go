//go:build !windows
// +build !windows

package iostreams

func (s *IOStreams) EnableVirtualTerminalProcessing() error {
	return nil
}
