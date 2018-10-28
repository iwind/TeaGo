package Tea

import (
	"github.com/iwind/TeaGo/assert"
	"os"
	"testing"
)

func TestFindLatestDir(t *testing.T) {
	t.Log(Root)
	t.Log(findLatestDir(Root, "public"))
}

func TestTmpDir(t *testing.T) {
	t.Log(TmpDir())
	t.Log(TmpFile("test.json"))
}

func TestIsTesting(t *testing.T) {
	a := assert.NewAssertion(t).Quiet()
	a.IsTrue(IsTesting())
	t.Log(os.Args)
}
