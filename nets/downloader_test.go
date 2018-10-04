package nets

import (
	"testing"
	"time"
	"github.com/iwind/TeaGo/utils/time"
	"github.com/iwind/TeaGo/logs"
)

func TestDownloaderItem_Start(t *testing.T) {
	var item = DownloaderItem{
		//url: "http://22018.xc.17yyba.com/xiaz/office2007%E5%AE%98%E6%96%B9%E4%B8%8B%E8%BD%BD%E5%85%8D%E8%B4%B9%E5%AE%8C%E6%95%B4%E7%89%88@418_3825.exe",
		url: "http://baidu.com/",
	}
	//item.target = "/Users/liuxiangchao/Documents/Projects/pp/apps/balelocal/src/main/tmp/test.exe"
	item.target = "/Users/liuxiangchao/Documents/Projects/pp/apps/balelocal/src/main/tmp/baidu.html"

	item.onStartFn = func() {
		t.Log("onstart")
	}
	item.onCompleteFn = func() {
		if item.success {
			t.Log("oncomplete success")
		} else {
			t.Log("oncomplete fail")
		}
	}
	item.onErrorFn = func() {
		t.Log("error:", item.error)
	}
	item.onProgressFn = func() {
		t.Logf("progress:%f, size:%d, at:%s.%d", item.progress, item.size, timeutil.Format("H:i:s"), time.Now().Nanosecond()/1000000)
	}
	item.Start()

	time.Sleep(5 * time.Second)
}

func TestDownloader_Start(t *testing.T) {
	var downloader = NewDownloader()
	downloader.OnStart(func(item *DownloaderItem) {
		logs.Println("start", item.tag)
	})
	downloader.OnCompleteFn(func(item *DownloaderItem) {
		logs.Println("complete", item.Tag(), item.Success())
	})
	downloader.OnAllCompleteFn(func() {
		logs.Println("all complete")
	})
	downloader.OnErrorFn(func(item *DownloaderItem) {
		logs.Println("error", item.Tag(), item.Error())
	})
	downloader.Add("http://22018.xc.17yyba.com/xiaz/office2007%E5%AE%98%E6%96%B9%E4%B8%8B%E8%BD%BD%E5%85%8D%E8%B4%B9%E5%AE%8C%E6%95%B4%E7%89%88@418_3825.exe", "id:1", "/Users/liuxiangchao/Documents/Projects/pp/apps/balelocal/src/main/tmp/test.exe")
	downloader.Add("http://22018.xc.17yyba.com/xiaz/office2007%E5%AE%98%E6%96%B9%E4%B8%8B%E8%BD%BD%E5%85%8D%E8%B4%B9%E5%AE%8C%E6%95%B4%E7%89%88@418_3825.exe", "id:2", "/Users/liuxiangchao/Documents/Projects/pp/apps/balelocal/src/main/tmp/test2.exe")
	downloader.Add("http://22018.xc.17yyba.com/xiaz/office2007%E5%AE%98%E6%96%B9%E4%B8%8B%E8%BD%BD%E5%85%8D%E8%B4%B9%E5%AE%8C%E6%95%B4%E7%89%88@418_3825.exe", "id:3", "/Users/liuxiangchao/Documents/Projects/pp/apps/balelocal/src/main/tmp/test3.exe")
	//downloader.Add("https://www.baidu.com/s?ie=utf-8&f=8&rsv_bp=0&rsv_idx=1&tn=baidu&wd=%E4%B8%8B%E8%BD%BD&rsv_pq=c8396baf0005c8fb&rsv_t=80d1%2F95jr8uHPxP0ogzhAFEqPG8CpWAaBmim4IDNU5hpX01D1roCrsrICCg&rqlang=cn&rsv_enter=1&rsv_sug3=7&rsv_sug1=2&rsv_sug7=100&rsv_sug2=0&inputT=3555&rsv_sug4=3555", "baidu", "/Users/liuxiangchao/Documents/Projects/pp/apps/balelocal/src/main/tmp/baidu.html")
	downloader.Start()

	time.Sleep(15 * time.Second)
}
