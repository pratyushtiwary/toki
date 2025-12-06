//go:build !windows

package process

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
)

type Process struct {
	ProcessBase
}

func (p *Process) Run(processGroupId *uintptr, env []string) error {
	cmd := exec.Command("/bin/sh", "-c", p.execCommand)
	cmd.Env = append(os.Environ(), env...)

	if processGroupId != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    int(*processGroupId),
		}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}
	p.cmd = cmd
	cmd.Stdout = p.stdoutBuffer
	cmd.Stderr = p.stderrBuffer
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return err
	}

	p.stdinWriter = &stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	if processGroupId != nil {
		p.pgId = *processGroupId
	} else {
		p.pgId = uintptr(cmd.Process.Pid)
	}

	return nil
}

func (p *Process) Suspend() error {
	err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGSTOP)

	if err != nil {
		return err
	}

	return nil
}

func (p *Process) Resume() error {
	err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGCONT)

	if err != nil {
		return err
	}

	return nil
}

func (p *Process) IsRunning() (bool, error) {
	if p.cmd == nil || p.cmd.Process == nil {
		return false, nil
	}

	err := p.cmd.Process.Signal(syscall.Signal(0))
	if err != nil {
		return false, nil
	}

	if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		return false, nil
	}

	return true, nil
}

func (p *Process) Cleanup() error {
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Signal(syscall.SIGINT)
		p.cmd.Wait()
		p.cmd.Process.Signal(syscall.SIGKILL)
	}
	return nil
}

func newCommand(execCommand string, parentProcessGroup *uintptr) (ProcessInterface, error) {
	base := ProcessBase{
		execCommand:  execCommand,
		stdoutBuffer: &bytes.Buffer{},
		stderrBuffer: &bytes.Buffer{},
	}

	if parentProcessGroup != nil {
		base.pgId = *parentProcessGroup
	}

	return &Process{ProcessBase: base}, nil
}

var NewCommand = newCommand
