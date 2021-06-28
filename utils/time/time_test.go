package timeutil

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	t.Log(Format("Y-m-d H:i:s"))
	t.Log(Format("Y-m-d H:i:s", time.Date(2020, 10, 10, 0, 0, 0, 0, time.Local)))
	t.Log(Format("c", time.Now().Add(-1*time.Hour)))
	t.Log(Format("r"))
	t.Log(Format("U"))
	t.Log(Format("D"))
	t.Log(Format("l"))
	t.Log(Format("A"))
	t.Log(Format("a"))
	t.Log(Format("F"))
	t.Log(Format("Y, y"))
	t.Log(Format("g, h"))
	t.Log(Format("u, v"))
	t.Log(Format("O, P"))
}

func BenchmarkFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Format("YmdHis.U")
	}
}
