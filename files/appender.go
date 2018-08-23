package files

import (
	"os"
	"sync"
)

type Appender struct {
	file   *os.File
	locker *sync.Mutex
}

func NewAppender(path string) (*Appender, error) {
	return NewFile(path).Appender()
}

func (this *Appender) AppendString(s string) (n int, err error) {
	return this.file.WriteString(s)
}

func (this *Appender) Append(b []byte) (n int, err error) {
	return this.file.Write(b)
}

func (this *Appender) Truncate(size ... int64) error {
	if len(size) > 0 {
		return this.file.Truncate(size[0])
	}
	return this.file.Truncate(0)
}

func (this *Appender) Sync() error {
	return this.file.Sync()
}

func (this *Appender) Lock() {
	this.locker.Lock()
}

func (this *Appender) Unlock() {
	this.locker.Unlock()
}

func (this *Appender) Close() error {
	return this.file.Close()
}
