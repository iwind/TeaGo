package stringutil

import (
	"log"
	"sync"
	"testing"
)

func TestRandString(t *testing.T) {
	t.Log(Rand(10))
}

func TestRandStringUnique(t *testing.T) {
	m := map[string]bool{}
	locker := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			s := Rand(16)
			locker.Lock()
			_, found := m[s]
			locker.Unlock()
			if found {
				log.Println("duplicated", s)
				return
			}
			locker.Lock()
			m[s] = true
			locker.Unlock()
		}()
	}
	wg.Wait()
	t.Log("all unique")
	t.Log(m)
}

func TestConvertID(t *testing.T) {
	t.Log(ConvertID(1234567890))
}

func TestVersionCompare(t *testing.T) {
	t.Log(VersionCompare("1.0", "1.0.3"))
	t.Log(VersionCompare("2.0.3", "2.0.3"))
	t.Log(VersionCompare("2", "2.1"))
	t.Log(VersionCompare("1.1.2", "1.2.1"))
	t.Log(VersionCompare("1.10.2", "1.2.1"))
	t.Log(VersionCompare("1.14.2", "1.1234567.1"))
}

func TestParseFileSize(t *testing.T) {
	{
		s, _ := ParseFileSize("1k")
		t.Logf("%f", s)
	}
	{
		s, _ := ParseFileSize("1m")
		t.Logf("%f", s)
	}
	{
		s, _ := ParseFileSize("1g")
		t.Logf("%f", s)
	}
}
