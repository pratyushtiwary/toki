package process

import (
	"bytes"
	"errors"
	"os/exec"
)

type ProcessBase struct {
	execCommand  string
	pgId         uintptr // in linux and darwin process group id is int, in WINDOWS IT IS UINTPTR!!!
	cmd          *exec.Cmd
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
}

func (cB *ProcessBase) Run(processGroupId *uintptr) error {
	return errors.New("don't directly call Run on base, extend and override")
}

func (cB *ProcessBase) Suspend() error {
	return errors.New("don't directly call Suspend on base, extend and override")
}

func (cB *ProcessBase) Resume() error {
	return errors.New("don't directly call Resume on base, extend and override")
}

func (c *ProcessBase) IsRunning() (bool, error) {
	return false, errors.New("don't directly call IsRunning on base, extend and override")
}

func (p *Process) WaitTillFinished() error {
	return p.cmd.Wait()
}

func (cB *ProcessBase) GetExecCommand() string {
	return cB.execCommand
}
func (cB *ProcessBase) GetPgId() uintptr {
	return cB.pgId
}
func (cB *ProcessBase) GetCmd() *exec.Cmd {
	return cB.cmd
}
func (cB *ProcessBase) GetStdoutBuffer() *bytes.Buffer {
	return cB.stdoutBuffer
}
func (cB *ProcessBase) GetStderrBuffer() *bytes.Buffer {
	return cB.stderrBuffer
}
