package files

import "github.com/iwind/TeaGo/Tea"

func NewTmpFile(file string) *File {
	return NewFile(Tea.TmpFile(file))
}
