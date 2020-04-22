package TeaGo

import (
	"bufio"
	"compress/gzip"
	"errors"
	"net"
	"net/http"
)

type gzipWriter struct {
	responseWriter http.ResponseWriter
	gzip           *gzip.Writer
}

func newGzipWriter(writer http.ResponseWriter, level int) (*gzipWriter, error) {
	w, err := gzip.NewWriterLevel(writer, level)
	if err != nil {
		return nil, err
	}

	writer.Header().Set("Content-Encoding", "gzip")
	writer.Header().Set("Transfer-Encoding", "chunked")
	writer.Header().Set("Vary", "Accept-Encoding")
	writer.Header().Set("Accept-encoding", "gzip, deflate, br")

	return &gzipWriter{
		responseWriter: writer,
		gzip:           w,
	}, nil
}

func (this *gzipWriter) Header() http.Header {
	return this.responseWriter.Header()
}

func (this *gzipWriter) WriteHeader(status int) {
	this.responseWriter.WriteHeader(status)
}

func (this *gzipWriter) Write(b []byte) (int, error) {
	return this.gzip.Write(b)
}

func (this *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := this.responseWriter.(http.Hijacker)
	if ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("http.Hijacker not implemented by underlying http.ResponseWriter")
}

func (this *gzipWriter) Close() error {
	_ = this.gzip.Flush()
	return this.gzip.Close()
}
