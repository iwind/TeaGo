package processes

import "testing"

func TestExecAndReturn(t *testing.T) {
	t.Log(Exec("/usr/local/bin/php", "-v"))
}
