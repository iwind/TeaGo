package files

import (
	"os"
	"time"
)

type Stat struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}
