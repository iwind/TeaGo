package files

type Whence = int

const (
	WhenceStart   = 0 // 文件开始处
	WhenceCurrent = 1 // 文件当前位置
	WhenceEnd     = 2 // 文件末尾处
)
