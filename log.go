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

func (this *DefaultLogWriter) Init() {
	this.queue = make(chan string, 10000)
	go func() {
		for {
			msg := <-this.queue
			log.Println(msg)
		}
	}()
}

func (this *DefaultLogWriter) Print(t time.Time, response *responseWriter, request *http.Request) {
	var tag = "ok"
	if response.status >= 400 {
		tag = "error"
	}

	this.queue <- logs.Sprintf("\n  <"+tag+">Request:\"%s %s %s\"</"+tag+">\n  RemoteAddr:%s\n  Status:%d\n  Bytes:%d\n  Referer:\"%s\"\n  UserAgent:\"%s\"\n  Cost:%.3fms",
		request.Method, request.RequestURI, request.Proto, request.RemoteAddr, response.status, response.bytes,
		request.Referer(), request.UserAgent(), float32(time.Since(t).Nanoseconds())/1000000)
}

func (this *DefaultLogWriter) Write(logMessage string) {
	this.queue <- logMessage
}

func (this *DefaultLogWriter) Close() {

}

type FileLogWriter struct {
	File       string
	fileWriter io.Writer
}

func (this *FileLogWriter) Init() {
	if len(this.File) == 0 {
		this.File = "logs/server.log"
	}
	file, err := os.OpenFile(this.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println("can not write file to 'logs/access.log':" + err.Error())
		return
	}
	this.fileWriter = file
}

func (this *FileLogWriter) Print(t time.Time, response *responseWriter, request *http.Request) {
	logs.Printf("\n  Request:\"%s %s %s\"\n  RemoteAddr:%s\n  Status:%d\n  Bytes:%d\n  Referer:\"%s\"\n  UserAgent:\"%s\"\n  Cost:%.3fms",
		request.Method, request.RequestURI, request.Proto, request.RemoteAddr, response.status, response.bytes,
		request.Referer(), request.UserAgent(), float32(time.Since(t).Nanoseconds())/1000000)
}

func (this *FileLogWriter) Write(logMessage string) {
	if this.fileWriter == nil {
		return
	}
	_, err := this.fileWriter.Write([]byte(logMessage))
	if err != nil {
		log.Println("Error:", err.Error())
	}
}

func (this *FileLogWriter) Close() {
	if this.fileWriter != nil {
		this.fileWriter.(*os.File).Close()
	}
}
