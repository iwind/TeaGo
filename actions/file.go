package actions

import (
	"mime/multipart"
	"io/ioutil"
	"io"
	"os"
	"github.com/iwind/TeaGo/files"
	"path/filepath"
)

// 上传的文件封装
type File struct {
	OriginFile *multipart.FileHeader

	Filename    string
	Size        int64
	Field       string
	Ext         string
	ContentType string
}

// 读取文件内容
func (this *File) Read() ([]byte, error) {
	multipartFile, err := this.OriginFile.Open()
	if err != nil {
		return nil, err
	}
	defer multipartFile.Close()
	return ioutil.ReadAll(multipartFile)
}

// 将文件内容写入到writer中
func (this *File) Write(writer io.Writer) (int, error) {
	fileBytes, err := this.Read()
	if err != nil {
		return 0, err
	}
	return writer.Write(fileBytes)
}

// 写入一个文件路径中
func (this *File) WriteToPath(path string) (int, error) {
	file := files.NewFile(filepath.Dir(path))
	if !file.Exists() {
		file.MkdirAll()
	}

	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	return this.Write(fp)
}
