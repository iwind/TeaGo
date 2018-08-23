package actions

import (
	"compress/gzip"
	"regexp"
	"github.com/iwind/TeaGo/types"
	"compress/flate"
	"github.com/iwind/TeaGo/logs"
)

type Gzip struct {
	Level int

	actionPtr  ActionWrapper
	gzipWriter *gzip.Writer
}

func (this *Gzip) BeforeAction(ptr ActionWrapper, paramName string) (goNext bool) {
	this.actionPtr = ptr
	ptr.Object().writer = this

	if this.Level <= 0 {
		reg := regexp.MustCompile("(\\d+)$")
		match := reg.FindAllString(paramName, 1)

		if len(match) == 0 {
			this.Level = gzip.DefaultCompression
		} else {
			level := types.Int(match[0])
			if level < flate.BestSpeed {
				level = flate.BestSpeed
			}
			if level > flate.BestCompression {
				level = flate.BestCompression
			}
			this.Level = level
		}
	}

	gzipWriter, err := gzip.NewWriterLevel(this.actionPtr.Object().ResponseWriter, this.Level)
	if err != nil {
		logs.Error(err)
	} else {
		this.gzipWriter = gzipWriter
	}

	return true
}

func (this *Gzip) Write(data []byte) (n int, err error) {
	var action = this.actionPtr.Object()

	if this.gzipWriter != nil {
		var header = action.ResponseWriter.Header()
		header.Set("Content-Encoding", "gzip")
		header.Set("Transfer-Encoding", "chunked")
		header.Set("Vary", "Accept-Encoding")
		header.Set("Accept-encoding", "gzip, deflate, br")

		n, err = this.gzipWriter.Write(data)
	} else {
		n, err = action.ResponseWriter.Write(data)
	}

	return
}

func (this *Gzip) AfterAction() {
	if this.gzipWriter != nil {
		this.gzipWriter.Close()
	}
}
