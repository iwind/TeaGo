// +build windows

package processes

import (
	"errors"
	"github.com/iwind/TeaGo/Tea"
	"os"
	"syscall"
)

type Process struct {
	dir     string
	command string
	args    []string
	native  *os.Process

	env []string

	in    *os.File
	out   *os.File
	err   *os.File
	files []*os.File

	pid int
}

func NewProcess(command string, args ...string) *Process {
	return &Process{
		command: command,
		args:    args,
	}
}

func (this *Process) SetPwd(dir string) {
	this.dir = dir
}

func (this *Process) SetIn(in *os.File) {
	this.in = in
}

func (this *Process) SetOut(out *os.File) {
	this.out = out
}

func (this *Process) SetErr(err *os.File) {
	this.err = err
}

// 添加共享的文件对象
func (this *Process) AppendFile(file ...*os.File) {
	this.files = append(this.files, file...)
}

// 添加环境变量
func (this *Process) AppendEnv(key, value string) {
	this.env = append(this.env, key+"="+value)
}

func (this *Process) Start() error {
	if this.in == nil {
		this.in = os.Stdin
	}

	if this.out == nil {
		this.out = os.Stdout
	}

	if this.err == nil {
		this.err = os.Stderr
	}

	files := []*os.File{this.in, this.out, this.err}
	if len(this.files) > 0 {
		files = append(files, this.files...)
	}

	pwd := Tea.Root
	if len(this.dir) > 0 {
		pwd = this.dir
	}

	// 环境变量
	env := os.Environ()
	if len(this.env) > 0 {
		env = append(env, this.env...)
	}

	attrs := os.ProcAttr{
		Dir:   pwd,
		Env:   env,
		Files: files,
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
	if this.in == nil {
		this.in = os.Stdin
	}

	if this.out == nil {
		this.out = os.Stdout
	}

	if this.err == nil {
		this.err = os.Stderr
	}

	files := []*os.File{this.in, this.out, this.err}
	if len(this.files) > 0 {
		files = append(files, this.files...)
	}

	pwd := Tea.Root
	if len(this.dir) > 0 {
		pwd = this.dir
	}

	// 环境变量
	env := os.Environ()
	if len(this.env) > 0 {
		env = append(env, this.env...)
	}

	attrs := os.ProcAttr{
		Dir:   pwd,
		Env:   env,
		Files: files,
		Sys: &syscall.SysProcAttr{
			HideWindow: true,
		},
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
