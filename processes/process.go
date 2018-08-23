// +build !windows

package processes

import (
	"os"
	"github.com/iwind/TeaGo/Tea"
	"errors"
	)

type Process struct {
	command string
	args    []string
	native  *os.Process
	out     *os.File
	pid     int
}

func NewProcess(command string, args ... string) *Process {
	return &Process{
		command: command,
		args:    args,
	}
}

func (process *Process) Out(out *os.File) {
	process.out = out
}

func (process *Process) Start() error {
	if process.out == nil {
		process.out = os.Stdout
	}

	attrs := os.ProcAttr{
		Dir:   Tea.Root,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, process.out, os.Stderr},
	}

	p, err := os.StartProcess(process.command, append([]string{process.command}, process.args ...), &attrs)
	if err != nil {
		return err
	}

	process.pid = p.Pid
	process.native = p
	return nil
}

func (process *Process) StartBackground() error {
	if process.out == nil {
		process.out = os.Stdout
	}

	attrs := os.ProcAttr{
		Dir:   Tea.Root,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, process.out, os.Stderr},
	}

	p, err := os.StartProcess(process.command, append([]string{process.command}, process.args ...), &attrs)
	if err != nil {
		return err
	}

	process.pid = p.Pid
	process.native = p
	return nil
}

func (process *Process) Wait() error {
	if process.native == nil {
		return errors.New("should not be start")
	}
	_, err := process.native.Wait()
	return err
}

func (process *Process) Pid() int {
	return process.pid
}
