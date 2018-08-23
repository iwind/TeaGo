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

func (downloader *downloader) Add(url string, tag string, target string) {
	downloader.items = append(downloader.items, &DownloaderItem{
		url:    url,
		tag:    tag,
		target: target,
	})
}

func (downloader *downloader) OnStart(fn func(item *DownloaderItem)) {
	downloader.onStartFn = fn
}

func (downloader *downloader) OnBeforeWriteFn(fn func(item *DownloaderItem)) {
	downloader.onBeforeWriteFn = fn
}

func (downloader *downloader) OnAfterWriteFn(fn func(item *DownloaderItem)) {
	downloader.onAfterWriteFn = fn
}

func (downloader *downloader) OnProgress(fn func(item *DownloaderItem)) {
	downloader.onProgressFn = fn
}

func (downloader *downloader) OnCompleteFn(fn func(item *DownloaderItem)) {
	downloader.onCompleteFn = fn
}

func (downloader *downloader) OnErrorFn(fn func(item *DownloaderItem)) {
	downloader.onErrorFn = fn
}

func (downloader *downloader) OnAllCompleteFn(fn func()) {
	downloader.onAllCompleteFn = fn
}

func (downloader *downloader) Concurrent(concurrent uint) {
	downloader.concurrent = concurrent
}

// 等待下载的条目
func (downloader *downloader) waitingItems() []*DownloaderItem {
	items := []*DownloaderItem{}
	for _, item := range downloader.items {
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

func (downloader *downloader) Start() {
	downloader.waitTasks(false)
}

func (downloader *downloader) Wait() {
	downloader.waitTasks(true)
}

func (downloader *downloader) waitTasks(loop bool) {
	go func() {
		// 放在循环中，以便支持动态添加的新的下载任务
		for {
			concurrent := downloader.concurrent
			if concurrent == 0 {
				concurrent = 1
			}

			// 是否有未完成的Items
			items := downloader.waitingItems()
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
						if downloader.onStartFn != nil {
							downloader.onStartFn(item)
						}
					}
					item.onErrorFn = func() {
						if downloader.onErrorFn != nil {
							downloader.onErrorFn(item)
						}
					}
					item.onCompleteFn = func() {
						wg.Add(-1)

						if downloader.onCompleteFn != nil {
							downloader.onCompleteFn(item)
						}
					}
					item.onBeforeWrite = func() {
						if downloader.onBeforeWriteFn != nil {
							downloader.onBeforeWriteFn(item)
						}
					}
					item.onAfterWrite = func() {
						if downloader.onAfterWriteFn != nil {
							downloader.onAfterWriteFn(item)
						}
					}
					item.onProgressFn = func() {
						if downloader.onProgressFn != nil {
							downloader.onProgressFn(item)
						}
					}
					item.Start()
				}(item)
			}

			// 等待此批任务完成
			wg.Wait()

			// 移除完成的任务
			leftItems := []*DownloaderItem{}
			for _, item := range downloader.items {
				if item.isCompleted {
					continue
				}
				leftItems = append(leftItems, item)
			}

			downloader.items = leftItems

			if len(leftItems) == 0 && downloader.onAllCompleteFn != nil {
				downloader.onAllCompleteFn()
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
