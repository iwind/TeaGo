package timers

import (
	"time"
	"sync"
)

// 延时执行函数，不阻塞当前进程
func Delay(duration time.Duration, task func(timer *time.Timer)) *time.Timer {
	timer := time.NewTimer(duration)
	go func() {
		<-timer.C
		task(timer)
	}()

	return timer
}

// 在某个时间点执行函数，不阻塞当前进程
func At(atTime time.Time, task func(timer *time.Timer)) *time.Timer {
	timer := time.NewTimer(-time.Since(atTime))
	go func() {
		<-timer.C
		task(timer)
	}()

	return timer
}

// 每隔一段时间执行函数，不阻塞当前进程
func Every(duration time.Duration, task func(ticker *time.Ticker)) *time.Ticker {
	ticker := time.NewTicker(duration)
	go func() {
		for {
			<-ticker.C
			task(ticker)
		}
	}()
	return ticker
}

// 循环执行某个函数，并保持每次执行之间的间隔
func Loop(duration time.Duration, task func(looper *Looper)) *Looper {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	looper := NewLooper()
	looper.wg = wg

	go func() {
		defer wg.Add(-1)

		for {
			if looper.isStopping {
				looper.isStopping = false
				return
			}

			task(looper)

			if looper.isStopping {
				looper.isStopping = false
				return
			}

			time.Sleep(duration)
		}
	}()

	return looper
}
