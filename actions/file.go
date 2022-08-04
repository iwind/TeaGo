package actions

import (
	"github.com/iwind/TeaGo/files"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// File 上传的文件封装
type File struct {
	OriginFile *multipart.FileHeader

	Filename    string
	Size        int64
	Field       string
	Ext         string
	ContentType string
}

func (this *File) Reader() (io.ReadCloser, error) {
	return this.OriginFile.Open()
}

// Read 读取文件内容
func (this *File) Read() ([]byte, error) {
	reader, err := this.OriginFile.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = reader.Close()
	}()
	return io.ReadAll(reader)
}

// WriteTo 将文件内容写入到writer中
func (this *File) WriteTo(writer io.Writer) (int64, error) {
	reader, err := this.Reader()
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = reader.Close()
	}()

	return io.Copy(writer, reader)
}

// WriteToPath 写入一个文件路径中
func (this *File) WriteToPath(path string) (int64, error) {
	var file = files.NewFile(filepath.Dir(path))
	if !file.Exists() {
		err := file.MkdirAll()
		if err != nil {
			return 0, err
		}
	}

	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = fp.Close()
	}()

	return this.WriteTo(fp)
}
