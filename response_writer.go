package TeaGo

import (
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

func (writer *responseWriter) Header() http.Header {
	return writer.responseWriter.Header()
}

func (writer *responseWriter) WriteHeader(status int) {
	writer.done = true
	writer.status = status
	writer.responseWriter.WriteHeader(status)
}

func (writer *responseWriter) Write(b []byte) (int, error) {
	writer.done = true
	length, err := writer.responseWriter.Write(b)
	if err == nil {
		writer.bytes += length
	}
	return length, err
}
