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

func (item *DownloaderItem) File() *os.File {
	return item.file
}

func (item *DownloaderItem) URL() string {
	return item.url
}

func (item *DownloaderItem) Tag() string {
	return item.tag
}

func (item *DownloaderItem) Target() string {
	return item.target
}

func (item *DownloaderItem) Progress() float32 {
	return item.progress
}

func (item *DownloaderItem) IsDownloading() bool {
	return item.isDownloading
}

func (item *DownloaderItem) IsCompleted() bool {
	return item.isCompleted
}

func (item *DownloaderItem) Success() bool {
	return item.success
}

func (item *DownloaderItem) ContentLength() uint64 {
	return item.contentLength
}

func (item *DownloaderItem) Size() uint64 {
	return item.size
}

func (item *DownloaderItem) Error() error {
	return item.error
}

func (item *DownloaderItem) Start() {
	defer func() {
		item.isDownloading = false
		item.isCompleted = true

		if item.success {
			stat, err := os.Stat(item.target)
			if err != nil {
				return
			}
			size := stat.Size()
			item.size = uint64(size)
			item.progress = float32(float64(item.size) / float64(item.contentLength))

			if item.onProgressFn != nil {
				item.onProgressFn()
			}
		}

		item.onCompleteFn()
	}()

	if item.onStartFn != nil {
		item.onStartFn()
	}

	// 是否已设置URL
	var url = strings.TrimSpace(item.url)
	if len(url) == 0 {
		item.addError(errors.New("'url' should be set before downloading"))
		return
	}

	// 是否已设置目标
	var target = strings.TrimSpace(item.target)
	if len(target) == 0 {
		item.addError(errors.New("'target' should be set before downloading"))
		return
	}

	if strings.HasSuffix(target, "/") || strings.HasSuffix(target, "\\") {
		target += filepath.Base(url)
	}

	// 初始化
	item.isDownloading = true
	item.isCompleted = false
	item.success = false

	var client = &http.Client{
		Timeout: 5 * time.Second,
	}
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		item.addError(err)
		return
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	resp, err := client.Do(request)
	if err != nil {
		item.addError(err)
		return
	}

	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		item.addError(err)
		return
	}
	item.contentLength = uint64(contentLength)

	file, err := os.OpenFile(item.target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		item.addError(err)
		return
	}
	item.file = file
	defer file.Close()

	// 如果内容长度为0
	if item.contentLength == 0 {
		if err != nil {
			item.addError(err)
			return
		}

		item.success = true
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
				stat, err := os.Stat(item.target)
				if err != nil {
					return
				}
				size := stat.Size()
				item.size = uint64(size)
				item.progress = float32(float64(item.size) / float64(item.contentLength))

				if item.onProgressFn != nil {
					item.onProgressFn()
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
		item.addError(err)
		return
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36 TeaDownloader/1.0")
	resp, err = client.Do(request)
	if err != nil {
		item.addError(err)
		resp.Body.Close()
		return
	}

	// 开始写入
	if item.onBeforeWrite != nil {
		item.onBeforeWrite()
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		item.addError(err)
		resp.Body.Close()
		return
	}

	if item.onAfterWrite != nil {
		item.onAfterWrite()
	}

	resp.Body.Close()
	item.success = true
}

func (item *DownloaderItem) addError(err error) {
	item.isDownloading = false
	item.isCompleted = true
	item.success = false
	item.error = err

	if item.onErrorFn != nil {
		item.onErrorFn()
	}
}
