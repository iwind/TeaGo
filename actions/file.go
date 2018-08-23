package actions

import (
	"mime/multipart"
	"io/ioutil"
	"io"
)

type File struct {
	OriginFile *multipart.FileHeader

	Filename    string
	Size        int64
	Field       string
	Ext         string
	ContentType string
}

func (this *File) Read() ([]byte, error) {
	multipartFile, err := this.OriginFile.Open()
	if err != nil {
		return nil, err
	}
	defer multipartFile.Close()
	return ioutil.ReadAll(multipartFile)
}

func (this *File) Write(writer io.Writer) (int, error) {
	fileBytes, err := this.Read()
	if err != nil {
		return 0, err
	}
	return writer.Write(fileBytes)
}
