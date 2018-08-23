package nets

import (
	"time"
	"sync"
)

type downloader struct {
	items           []*DownloaderItem
	concurrent      uint
	onStartFn       func(item *DownloaderItem)
	onBeforeWriteFn func(item *DownloaderItem)
	onAfterWriteFn  func(item *DownloaderItem)
	onProgressFn    func(item *DownloaderItem)

	onCompleteFn    func(item *DownloaderItem)
	onErrorFn       func(item *DownloaderItem)
	onAllCompleteFn func()
}

func NewDownloader() *downloader {
	return &downloader{
		items:      []*DownloaderItem{},
		concurrent: 1,
	}
}

func (this *downloader) Add(url string, tag string, target string) {
	this.items = append(this.items, &DownloaderItem{
		url:    url,
		tag:    tag,
		target: target,
	})
}

func (this *downloader) OnStart(fn func(item *DownloaderItem)) {
	this.onStartFn = fn
}

func (this *downloader) OnBeforeWriteFn(fn func(item *DownloaderItem)) {
	this.onBeforeWriteFn = fn
}

func (this *downloader) OnAfterWriteFn(fn func(item *DownloaderItem)) {
	this.onAfterWriteFn = fn
}

func (this *downloader) OnProgress(fn func(item *DownloaderItem)) {
	this.onProgressFn = fn
}

func (this *downloader) OnCompleteFn(fn func(item *DownloaderItem)) {
	this.onCompleteFn = fn
}

func (this *downloader) OnErrorFn(fn func(item *DownloaderItem)) {
	this.onErrorFn = fn
}

func (this *downloader) OnAllCompleteFn(fn func()) {
	this.onAllCompleteFn = fn
}

func (this *downloader) Concurrent(concurrent uint) {
	this.concurrent = concurrent
}

// 等待下载的条目
func (this *downloader) waitingItems() []*DownloaderItem {
	items := []*DownloaderItem{}
	for _, item := range this.items {
		if item.isDownloading {
			continue
		}
		if item.isCompleted {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (this *downloader) Start() {
	this.waitTasks(false)
}

func (this *downloader) Wait() {
	this.waitTasks(true)
}

func (this *downloader) waitTasks(loop bool) {
	go func() {
		// 放在循环中，以便支持动态添加的新的下载任务
		for {
			concurrent := this.concurrent
			if concurrent == 0 {
				concurrent = 1
			}

			// 是否有未完成的Items
			items := this.waitingItems()
			countItems := uint(len(items))
			if countItems == 0 {
				if !loop {
					return
				}

				time.Sleep(1 * time.Second)
				continue
			}

			if countItems < concurrent {
				concurrent = countItems
			} else {
				items = items[:concurrent]
			}

			wg := &sync.WaitGroup{}
			wg.Add(int(concurrent))

			for _, item := range items {
				go func(item *DownloaderItem) {
					item.onStartFn = func() {
						if this.onStartFn != nil {
							this.onStartFn(item)
						}
					}
					item.onErrorFn = func() {
						if this.onErrorFn != nil {
							this.onErrorFn(item)
						}
					}
					item.onCompleteFn = func() {
						wg.Add(-1)

						if this.onCompleteFn != nil {
							this.onCompleteFn(item)
						}
					}
					item.onBeforeWrite = func() {
						if this.onBeforeWriteFn != nil {
							this.onBeforeWriteFn(item)
						}
					}
					item.onAfterWrite = func() {
						if this.onAfterWriteFn != nil {
							this.onAfterWriteFn(item)
						}
					}
					item.onProgressFn = func() {
						if this.onProgressFn != nil {
							this.onProgressFn(item)
						}
					}
					item.Start()
				}(item)
			}

			// 等待此批任务完成
			wg.Wait()

			// 移除完成的任务
			leftItems := []*DownloaderItem{}
			for _, item := range this.items {
				if item.isCompleted {
					continue
				}
				leftItems = append(leftItems, item)
			}

			this.items = leftItems

			if len(leftItems) == 0 && this.onAllCompleteFn != nil {
				this.onAllCompleteFn()
			}
		}
	}()
}

/**func (downloader *downloader) Pause() {
	//@TODO 需要实现
	logs.Errorf("downloader.Pause(): %s", "此方法等待实现")
}

func (downloader *downloader) Stop() {
	//@TODO 需要实现
	logs.Errorf("downloader.Stop(): %s", "此方法等待实现")
}**/
