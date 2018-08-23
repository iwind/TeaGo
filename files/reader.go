package files

import (
	"os"
	"io"
	"github.com/iwind/TeaGo/logs"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/go-yaml/yaml"
)

type Reader struct {
	file *os.File
}

func NewReader(path string) (*Reader, error) {
	return NewFile(path).Reader()
}

func (reader *Reader) Read(size int64) []byte {
	data := make([]byte, size)
	n, err := reader.file.Read(data)
	if err != nil {
		if err != io.EOF {
			logs.Error(err)
		} else {
			return []byte{}
		}
	}
	if int64(n) < size {
		data = data[:n]
	}
	return data
}

func (reader *Reader) ReadByte() []byte {
	return reader.Read(1)
}

func (reader *Reader) ReadLine() []byte {
	line := []byte{}
	for {
		b := reader.ReadByte()
		if len(b) == 0 {
			return line
		}

		line = append(line, b[0])
		if b[0] == '\n' || b[0] == '\r' {
			break
		}
	}
	return line
}

func (reader *Reader) ReadAll() []byte {
	stat, err := reader.file.Stat()
	if err != nil {
		logs.Error(err)
		return []byte{}
	}

	return reader.Read(stat.Size())
}

func (reader *Reader) ReadJSON(ptr interface{}) error {
	data := reader.ReadAll()
	return ffjson.Unmarshal(data, ptr)
}

func (reader *Reader) ReadYAML(ptr interface{}) error {
	data := reader.ReadAll()
	return yaml.Unmarshal(data, ptr)
}

func (reader *Reader) Seek(offset int64, whence ... int) (ret int64, err error) {
	if len(whence) > 0 {
		return reader.file.Seek(offset, whence[0])
	}
	return reader.file.Seek(offset, 0)
}

func (reader *Reader) Reset() error {
	_, err := reader.Seek(0)
	return err
}

func (reader *Reader) Length() (length int64, err error) {
	stat, err := reader.file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

func (reader *Reader) Close() error {
	return reader.file.Close()
}
