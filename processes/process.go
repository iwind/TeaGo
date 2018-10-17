// +build !windows

package processes

import (
	"errors"
	"github.com/iwind/TeaGo/Tea"
	"os"
)

type Process struct {
	dir     string
	command string
	args    []string
	native  *os.Process
	out     *os.File
	pid     int
}

func NewProcess(command string, args ...string) *Process {
	return &Process{
		command: command,
		args:    args,
	}
}

func (this *Process) Pwd(dir string) {
	this.dir = dir
}

func (this *Process) Out(out *os.File) {
	this.out = out
}

func (this *Process) Start() error {
	if this.out == nil {
		this.out = os.Stdout
	}

	pwd := Tea.Root
	if len(this.dir) > 0 {
		pwd = this.dir
	}
	attrs := os.ProcAttr{
		Dir:   pwd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, this.out, os.Stderr},
	}

	p, err := os.StartProcess(this.command, append([]string{this.command}, this.args...), &attrs)
	if err != nil {
		return err
	}

	this.pid = p.Pid
	this.native = p
	return nil
}

func (this *Process) StartBackground() error {
	if this.out == nil {
		this.out = os.Stdout
	}

	pwd := Tea.Root
	if len(this.dir) > 0 {
		pwd = this.dir
	}

	attrs := os.ProcAttr{
		Dir:   pwd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, this.out, os.Stderr},
	}

	p, err := os.StartProcess(this.command, append([]string{this.command}, this.args...), &attrs)
	if err != nil {
		return err
	}

	this.pid = p.Pid
	this.native = p
	return nil
}

func (this *Process) Wait() error {
	if this.native == nil {
		return errors.New("should not be start")
	}
	_, err := this.native.Wait()
	return err
}

func (this *Process) Pid() int {
	return this.pid
}
