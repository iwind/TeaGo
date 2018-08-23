package TeaGo

import (
	"time"
	"net/http"
	"github.com/iwind/TeaGo/logs"
	"io"
	"os"
	"log"
)

type LogWriter interface {
	Init()
	Print(t time.Time, response *responseWriter, request *http.Request)
	Write(logMessage string)
	Close()
}

type DefaultLogWriter struct {
	queue chan string
}

func (writer *DefaultLogWriter) Init() {
	writer.queue = make(chan string, 10000)
	go func() {
		for {
			msg := <- writer.queue
			log.Println(msg)
		}
	}()
}

func (writer *DefaultLogWriter) Print(t time.Time, response *responseWriter, request *http.Request) {
	var tag = "ok"
	if response.status >= 400 {
		tag = "error"
	}

	writer.queue <- logs.Sprintf("\n  <"+tag+">Request:\"%s %s %s\"</"+tag+">\n  RemoteAddr:%s\n  Status:%d\n  Bytes:%d\n  Referer:\"%s\"\n  UserAgent:\"%s\"\n  Cost:%.3fms",
		request.Method, request.RequestURI, request.Proto, request.RemoteAddr, response.status, response.bytes,
		request.Referer(), request.UserAgent(), float32(time.Since(t).Nanoseconds())/1000000)
}

func (writer *DefaultLogWriter) Write(logMessage string) {
	writer.queue <- logMessage
}

func (writer *DefaultLogWriter) Close() {

}

type FileLogWriter struct {
	File       string
	fileWriter io.Writer
}

func (writer *FileLogWriter) Init() {
	if len(writer.File) == 0 {
		writer.File = "logs/server.log"
	}
	file, err := os.OpenFile(writer.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println("can not write file to 'logs/access.log':" + err.Error())
		return
	}
	writer.fileWriter = file
}

func (writer *FileLogWriter) Print(t time.Time, response *responseWriter, request *http.Request) {
	logs.Printf("\n  Request:\"%s %s %s\"\n  RemoteAddr:%s\n  Status:%d\n  Bytes:%d\n  Referer:\"%s\"\n  UserAgent:\"%s\"\n  Cost:%.3fms",
		request.Method, request.RequestURI, request.Proto, request.RemoteAddr, response.status, response.bytes,
		request.Referer(), request.UserAgent(), float32(time.Since(t).Nanoseconds())/1000000)
}

func (writer *FileLogWriter) Write(logMessage string) {
	if writer.fileWriter == nil {
		return
	}
	_, err := writer.fileWriter.Write([]byte(logMessage))
	if err != nil {
		log.Println("Error:", err.Error())
	}
}

func (writer *FileLogWriter) Close() {
	if writer.fileWriter != nil {
		writer.fileWriter.(*os.File).Close()
	}
}
