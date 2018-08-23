package Tea

import "testing"

func TestFindLatestDir(t *testing.T) {
	t.Log(Root)
	t.Log(findLatestDir(Root, "public"))
}

func TestTmpDir(t *testing.T) {
	t.Log(TmpDir())
	t.Log(TmpFile("test.json"))
}
