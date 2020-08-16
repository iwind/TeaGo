package files

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

type Reader struct {
	file *os.File
}

func NewReader(path string) (*Reader, error) {
	return NewFile(path).Reader()
}

func (this *Reader) Read(size int64) []byte {
	data := make([]byte, size)
	n, err := this.file.Read(data)
	if err != nil {
		if err != io.EOF {
			logs.Error(err)
		}
	}
	if n > 0 {
		return data[:n]
	}
	return []byte{}
}

func (this *Reader) ReadByte() []byte {
	return this.Read(1)
}

func (this *Reader) ReadLine() []byte {
	line := []byte{}
	for {
		b := this.ReadByte()
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

func (this *Reader) ReadAll() []byte {
	stat, err := this.file.Stat()
	if err != nil {
		logs.Error(err)
		return []byte{}
	}

	return this.Read(stat.Size())
}

// 从JSON文件中读取数据
func (this *Reader) ReadJSON(ptr interface{}) error {
	data := this.ReadAll()
	return json.Unmarshal(data, ptr)
}

// 从JSON文件中读取Map数据
func (this *Reader) ReadJSONMap() (maps.Map, error) {
	m := maps.Map{}
	err := this.ReadJSON(&m)
	return m, err
}

// 从YAML文件中读取数据
func (this *Reader) ReadYAML(ptr interface{}) error {
	data := this.ReadAll()
	return yaml.Unmarshal(data, ptr)
}

// 从YAML文件中读取Map数据
func (this *Reader) ReadYAMLMap() (maps.Map, error) {
	m := maps.Map{}
	err := this.ReadYAML(&m)
	return m, err
}

// 移动文件指针到某个位置
func (this *Reader) Seek(offset int64, whence ...Whence) (ret int64, err error) {
	if len(whence) > 0 {
		return this.file.Seek(offset, whence[0])
	}
	return this.file.Seek(offset, 0)
}

func (this *Reader) Reset() error {
	_, err := this.Seek(0)
	return err
}

func (this *Reader) Length() (length int64, err error) {
	stat, err := this.file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

func (this *Reader) Close() error {
	return this.file.Close()
}
