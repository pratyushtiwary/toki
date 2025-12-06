//go:build windows

package process

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

type Process struct {
	ProcessBase
	job windows.Handle
}

var (
	ntdll            = windows.NewLazySystemDLL("ntdll.dll")
	ntSuspendProcess = ntdll.NewProc("NtSuspendProcess")
	ntResumeProcess  = ntdll.NewProc("NtResumeProcess")
)

func (p *Process) Run(processGroupId *uintptr, env []string) error {
	cmd := exec.Command("cmd.exe", "/C", p.execCommand)
	cmd.Env = append(os.Environ(), env...)
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

	processHandle, err := windows.OpenProcess(
		windows.PROCESS_ALL_ACCESS,
		false,
		uint32(cmd.Process.Pid),
	)

	if err != nil {
		return err
	}

	var job windows.Handle

	if processGroupId != nil {
		job = windows.Handle(*processGroupId)
	} else {
		job = p.job
	}

	p.pgId = uintptr(job)
	if err := windows.AssignProcessToJobObject(job, processHandle); err != nil {
		return err
	}

	return nil
}

func (p *Process) Suspend() error {
	processHandle, err := windows.OpenProcess(
		windows.PROCESS_ALL_ACCESS,
		false,
		uint32(p.cmd.Process.Pid),
	)

	if err != nil {
		return err
	}

	return1, _, err := ntSuspendProcess.Call(uintptr(processHandle))

	if return1 != 0 {
		return err
	}

	return nil
}

func (p *Process) Resume() error {
	processHandle, err := windows.OpenProcess(
		windows.PROCESS_ALL_ACCESS,
		false,
		uint32(p.cmd.Process.Pid),
	)

	if err != nil {
		return err
	}

	return1, _, err := ntResumeProcess.Call(uintptr(processHandle))

	if return1 != 0 {
		return err
	}

	return nil
}

func (p *Process) IsRunning() (bool, error) {
	processHandle, err := windows.OpenProcess(
		windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		uint32(p.cmd.Process.Pid),
	)

	if err != nil {
		return false, err
	}

	status, err := windows.WaitForSingleObject(processHandle, 0)
	if err != nil {
		return false, err
	}

	return status == uint32(windows.WAIT_TIMEOUT), nil
}

func (p *Process) Cleanup() error {
	if p.cmd != nil && p.cmd.Process != nil {
		err := p.cmd.Process.Kill()
		if err != nil {
			if !strings.Contains(err.Error(), "already finished") {
				return err
			}
		}
		p.cmd.Process.Release()
	}

	if p.job != 0 {
		err := windows.CloseHandle(p.job)
		if err != nil {
			return err
		}
		p.job = 0
	}

	return nil
}

func newCommand(execCommand string, parentProcessGroup *uintptr) (ProcessInterface, error) {
	var job windows.Handle
	var err error

	if parentProcessGroup != nil {
		job = windows.Handle(*parentProcessGroup)
	} else {
		job, err = windows.CreateJobObject(nil, nil)

		if err != nil {
			return nil, err
		}
	}

	base := ProcessBase{
		execCommand:  execCommand,
		stdoutBuffer: &bytes.Buffer{},
		stderrBuffer: &bytes.Buffer{},
	}

	return &Process{
		job:         job,
		ProcessBase: base,
	}, nil
}

var NewCommand = newCommand
