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

func (appender *Appender) AppendString(s string) (n int, err error) {
	return appender.file.WriteString(s)
}

func (appender *Appender) Append(b []byte) (n int, err error) {
	return appender.file.Write(b)
}

func (appender *Appender) Truncate(size ... int64) error {
	if len(size) > 0 {
		return appender.file.Truncate(size[0])
	}
	return appender.file.Truncate(0)
}

func (appender *Appender) Sync() error {
	return appender.file.Sync()
}

func (appender *Appender) Lock() {
	appender.locker.Lock()
}

func (appender *Appender) Unlock() {
	appender.locker.Unlock()
}

func (appender *Appender) Close() error {
	return appender.file.Close()
}
