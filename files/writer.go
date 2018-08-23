package files

import (
	"os"
	"sync"
	"io"
)

type Writer struct {
	file   *os.File
	locker *sync.Mutex
}

func NewWriter(path string) (*Writer, error) {
	return NewFile(path).Writer()
}

func (writer *Writer) WriteString(s string) (n int64, err error) {
	written, err := writer.file.WriteString(s)
	n = int64(written)
	return
}

func (writer *Writer) Write(b []byte) (n int64, err error) {
	written, err := writer.file.Write(b)
	n = int64(written)
	return
}

func (writer *Writer) WriteIOReader(reader io.Reader) (n int64, err error) {
	return io.Copy(writer.file, reader)
}

func (writer *Writer) Truncate(size ... int64) error {
	if len(size) > 0 {
		return writer.file.Truncate(size[0])
	}
	return writer.file.Truncate(0)
}

func (writer *Writer) Seek(offset int64, whence ... int) (ret int64, err error) {
	if len(whence) > 0 {
		return writer.file.Seek(offset, whence[0])
	}
	return writer.file.Seek(offset, 0)
}

func (writer *Writer) Sync() error {
	return writer.file.Sync()
}

func (writer *Writer) Lock() {
	writer.locker.Lock()
}

func (writer *Writer) Unlock() {
	writer.locker.Unlock()
}

func (writer *Writer) Close() error {
	return writer.file.Close()
}
