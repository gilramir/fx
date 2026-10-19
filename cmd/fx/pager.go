package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// pager is a running pager process that reads what we write to it.
type pager struct {
	cmd *exec.Cmd
	in  io.WriteCloser
}

// pagerCommand returns the shell command to page with: $PAGER if set,
// else less, else more.
func pagerCommand() string {
	if p := os.Getenv("PAGER"); p != "" {
		return p
	}
	if _, err := exec.LookPath("less"); err == nil {
		return "less"
	}
	return "more"
}

// startPager runs the pager with its output on our stdout. $PAGER may
// contain arguments, so it is run through the shell, as git and man do.
func startPager() (*pager, error) {
	cmd := exec.Command("/bin/sh", "-c", pagerCommand())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// less: quit at once if everything fits on one screen, and don't
	// clear the screen on exit, unless the user configured it otherwise.
	if os.Getenv("LESS") == "" {
		cmd.Env = append(os.Environ(), "LESS=FRX")
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &pager{cmd: cmd, in: in}, nil
}

// finish closes the pager's input and waits for the user to quit it.
// writeErr is the error from writing the dump; if the user quit the pager
// early, writes fail with EPIPE, which is not an error for us.
func (p *pager) finish(writeErr error) error {
	p.in.Close()
	waitErr := p.cmd.Wait()
	if writeErr != nil && !errors.Is(writeErr, syscall.EPIPE) {
		return writeErr
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		return nil // the pager reports its own problems
	}
	return waitErr
}
