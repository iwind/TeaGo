package nets

import (
	"net/http"
	"strconv"
	"os"
	"errors"
	"strings"
	"io"
	"time"
	"path/filepath"
)

type DownloaderItem struct {
	url           string
	tag           string
	target        string
	progress      float32
	isDownloading bool
	isCompleted   bool
	contentLength uint64
	size          uint64
	error         error
	success       bool
	file          *os.File

	onStartFn     func()
	onBeforeWrite func()
	onAfterWrite  func()
	onProgressFn  func()
	onCompleteFn  func()
	onErrorFn     func()
}

func (this *DownloaderItem) File() *os.File {
	return this.file
}

func (this *DownloaderItem) URL() string {
	return this.url
}

func (this *DownloaderItem) Tag() string {
	return this.tag
}

func (this *DownloaderItem) Target() string {
	return this.target
}

func (this *DownloaderItem) Progress() float32 {
	return this.progress
}

func (this *DownloaderItem) IsDownloading() bool {
	return this.isDownloading
}

func (this *DownloaderItem) IsCompleted() bool {
	return this.isCompleted
}

func (this *DownloaderItem) Success() bool {
	return this.success
}

func (this *DownloaderItem) ContentLength() uint64 {
	return this.contentLength
}

func (this *DownloaderItem) Size() uint64 {
	return this.size
}

func (this *DownloaderItem) Error() error {
	return this.error
}

func (this *DownloaderItem) Start() {
	defer func() {
		this.isDownloading = false
		this.isCompleted = true

		if this.success {
			stat, err := os.Stat(this.target)
			if err != nil {
				return
			}
			size := stat.Size()
			this.size = uint64(size)
			this.progress = float32(float64(this.size) / float64(this.contentLength))

			if this.onProgressFn != nil {
				this.onProgressFn()
			}
		}

		this.onCompleteFn()
	}()

	if this.onStartFn != nil {
		this.onStartFn()
	}

	// 是否已设置URL
	var url = strings.TrimSpace(this.url)
	if len(url) == 0 {
		this.addError(errors.New("'url' should be set before downloading"))
		return
	}

	// 是否已设置目标
	var target = strings.TrimSpace(this.target)
	if len(target) == 0 {
		this.addError(errors.New("'target' should be set before downloading"))
		return
	}

	if strings.HasSuffix(target, "/") || strings.HasSuffix(target, "\\") {
		target += filepath.Base(url)
	}

	// 初始化
	this.isDownloading = true
	this.isCompleted = false
	this.success = false

	var client = &http.Client{
		Timeout: 5 * time.Second,
	}
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		this.addError(err)
		return
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	resp, err := client.Do(request)
	if err != nil {
		this.addError(err)
		return
	}

	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		this.addError(err)
		return
	}
	this.contentLength = uint64(contentLength)

	file, err := os.OpenFile(this.target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		this.addError(err)
		return
	}
	this.file = file
	defer file.Close()

	// 如果内容长度为0
	if this.contentLength == 0 {
		if err != nil {
			this.addError(err)
			return
		}

		this.success = true
		return
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	tickerDone := make(chan bool)
	go func() {
		for {
			select {
			case <-tickerDone:
				return
			case <-ticker.C:
				stat, err := os.Stat(this.target)
				if err != nil {
					return
				}
				size := stat.Size()
				this.size = uint64(size)
				this.progress = float32(float64(this.size) / float64(this.contentLength))

				if this.onProgressFn != nil {
					this.onProgressFn()
				}
			}
		}
	}()
	defer func() {
		ticker.Stop()
		tickerDone <- true
	}()

	// 开始下载
	client = &http.Client{}
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		this.addError(err)
		return
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36 TeaDownloader/1.0")
	resp, err = client.Do(request)
	if err != nil {
		this.addError(err)
		resp.Body.Close()
		return
	}

	// 开始写入
	if this.onBeforeWrite != nil {
		this.onBeforeWrite()
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		this.addError(err)
		resp.Body.Close()
		return
	}

	if this.onAfterWrite != nil {
		this.onAfterWrite()
	}

	resp.Body.Close()
	this.success = true
}

func (this *DownloaderItem) addError(err error) {
	this.isDownloading = false
	this.isCompleted = true
	this.success = false
	this.error = err

	if this.onErrorFn != nil {
		this.onErrorFn()
	}
}
