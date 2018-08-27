package actions

import (
	"net/http"
)

type TestingResponseWriter struct {
	Data []byte
}

func (this *TestingResponseWriter) Header() http.Header {
	return http.Header{}
}

func (this *TestingResponseWriter) Write(data []byte) (int, error) {
	this.Data = append(this.Data, data ...)
	return len(data), nil
}

func (this *TestingResponseWriter) WriteHeader(int) {

}
