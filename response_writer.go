package TeaGo

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type responseWriter struct {
	responseWriter http.ResponseWriter
	status         int
	bytes          int
	done           bool
}

func newResponseWriter(writer http.ResponseWriter) *responseWriter {
	return &responseWriter{
		responseWriter: writer,
		status:         200,
	}
}

func (this *responseWriter) Header() http.Header {
	return this.responseWriter.Header()
}

func (this *responseWriter) WriteHeader(status int) {
	this.done = true
	this.status = status
	this.responseWriter.WriteHeader(status)
}

func (this *responseWriter) Write(b []byte) (int, error) {
	this.done = true
	length, err := this.responseWriter.Write(b)
	if err == nil {
		this.bytes += length
	}
	return length, err
}

func (this *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := this.responseWriter.(http.Hijacker)
	if ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("http.Hijacker not implemented by underlying http.ResponseWriter")
}
