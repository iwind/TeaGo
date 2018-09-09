package files

import (
	"os"
	"path/filepath"
	"time"
	"io/ioutil"
	"github.com/iwind/TeaGo/utils/string"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/Tea"
	"sync"
	"fmt"
)

// 文件对象定义
type File struct {
	path string
}

// 包装新文件对象
func NewFile(path string) *File {
	return &File{
		path: path,
	}
}

// 取得文件名
func (this *File) Name() string {
	return filepath.Base(this.path)
}

// 取得文件统计信息
func (this *File) Stat() (*Stat, error) {
	stat, err := os.Stat(this.path)
	if err != nil {
		return nil, err
	}

	return &Stat{
		Name:    stat.Name(),
		Size:    stat.Size(),
		Mode:    stat.Mode(),
		ModTime: stat.ModTime(),
		IsDir:   stat.IsDir(),
	}, err
}

// 判断文件是否存在
func (this *File) Exists() bool {
	_, err := os.Stat(this.path)
	return !os.IsNotExist(err)
}

// 取得父级目录对象
func (this *File) Parent() *File {
	return NewFile(filepath.Dir(this.path))
}

// 判断是否为目录
func (this *File) IsDir() bool {
	stat, err := this.Stat()
	if err != nil {
		return false
	}

	return stat.IsDir
}

// 判断是否为文件
func (this *File) IsFile() bool {
	stat, err := this.Stat()
	if err != nil {
		return false
	}

	return stat.Mode.IsRegular()
}

// 读取最后更新时间
func (this *File) LastModified() (time.Time, error) {
	stat, err := this.Stat()
	if err != nil {
		return time.Unix(0, 0), err
	}
	return stat.ModTime, nil
}

// 读取文件尺寸
func (this *File) Size() (int64, error) {
	stat, err := this.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size, nil
}

// 读取文件模式
func (this *File) Mode() (os.FileMode, error) {
	stat, err := this.Stat()
	if err != nil {
		return os.FileMode(0), err
	}
	return stat.Mode, nil
}

// 读取文件路径
func (this *File) Path() string {
	return this.path
}

// 读取文件绝对路径
func (this *File) AbsPath() (string, error) {
	p, err := filepath.Abs(this.path)
	if err != nil {
		return p, err
	}
	return p, nil
}

// 对文件内容进行Md5处理
func (this *File) Md5() (string, error) {
	data, err := ioutil.ReadFile(this.path)
	if err != nil {
		return "", err
	}
	return stringutil.Md5(string(data)), err
}

// 读取文件内容
func (this *File) ReadAll() ([]byte, error) {
	data, err := ioutil.ReadFile(this.path)
	if err != nil {
		return []byte{}, err
	}
	return data, err
}

// 读取文件内容并返回字符串形式
func (this *File) ReadAllString() (string, error) {
	data, err := ioutil.ReadFile(this.path)
	if err != nil {
		return "", err
	}
	return string(data), err
}

// 写入数据
func (this *File) Write(data []byte) error {
	writer, err := this.Writer()
	if err != nil {
		return err
	}
	writer.Lock()
	writer.Truncate()
	_, err = writer.Write(data)
	writer.Unlock()
	writer.Close()
	return err
}

// 写入字符串数据
func (this *File) WriteString(data string) error {
	return this.Write([]byte(data))
}

// 写入格式化的字符串数据
func (this *File) WriteFormat(format string, args ... interface{}) error {
	return this.WriteString(fmt.Sprintf(format, args ...))
}

// 在文件末尾写入数据
func (this *File) Append(data []byte) error {
	appender, err := this.Appender()
	if err != nil {
		return err
	}
	appender.Lock()
	_, err = appender.Append(data)
	appender.Unlock()
	appender.Close()
	return err
}

// 在文件末尾写入字符串数据
func (this *File) AppendString(data string) error {
	return this.Append([]byte(data))
}

// 取得文件扩展名，带点符号
func (this *File) Ext() string {
	return filepath.Ext(this.path)
}

// 目录下的文件
func (this *File) Child(filename string) *File {
	return NewFile(this.path + Tea.DS + filename)
}

// 列出目录下级的子文件对象
// 注意只会返回下一级，不会递归深入子目录
func (this *File) List() []*File {
	result := []*File{}

	if !this.IsDir() {
		return result
	}

	path, err := this.AbsPath()
	if err != nil {
		logs.Error(err)
		return result
	}

	fp, err := os.OpenFile(this.path, os.O_RDONLY, 0444)
	if err != nil {
		logs.Error(err)
		return result
	}

	defer fp.Close()
	names, err := fp.Readdirnames(-1)
	if err != nil {
		logs.Error(err)
		return result
	}

	for _, name := range names {
		result = append(result, NewFile(path+Tea.DS+name))
	}

	return result
}

// 使用模式匹配查找当前目录下的文件
func (this *File) Glob(pattern string) []*File {
	result := []*File{}
	matches, err := filepath.Glob(this.path + Tea.DS + pattern)
	if err != nil {
		return result
	}

	for _, path := range matches {
		result = append(result, NewFile(path))
	}

	return result
}

// 递归地对当前目录下的所有子文件、目录应用迭代器
func (this *File) Range(iterator func(file *File)) {
	if !this.Exists() || !this.IsDir() {
		return
	}

	for _, childFile := range this.List() {
		if childFile.IsDir() {
			childFile.Range(iterator)
		} else {
			iterator(childFile)
		}
	}
}

// 创建目录，但如果父级目录不存在，则会失败
func (this *File) Mkdir(perm ... os.FileMode) error {
	if len(perm) > 0 {
		return os.Mkdir(this.path, perm[0])
	}
	return os.Mkdir(this.path, 0777)
}

// 创建多级目录
func (this *File) MkdirAll(perm ... os.FileMode) error {
	if len(perm) > 0 {
		return os.MkdirAll(this.path, perm[0])
	}
	return os.MkdirAll(this.path, 0777)
}

// 创建文件
func (this *File) Create() error {
	fp, err := os.Create(this.path)
	if err != nil {
		return err
	}
	fp.Close()
	return nil
}

// 修改文件的访问和修改时间为当前时间
func (this *File) Touch() error {
	now := time.Now()
	return os.Chtimes(this.path, now, now)
}

// 删除文件或目录，但如果目录不为空则会失败
func (this *File) Delete() error {
	if !this.IsDir() {
		return os.Remove(this.path)
	}
	return os.Remove(this.path)
}

// 判断文件或目录是否存在，然后删除文件或目录，如果目录不为空则会失败
func (this *File) DeleteIfExists() error {
	if !this.Exists() {
		return nil
	}
	if !this.IsDir() {
		return os.Remove(this.path)
	}
	return os.Remove(this.path)
}

// 删除文件或目录，即使目录不为空也会删除
func (this *File) DeleteAll() error {
	if !this.IsDir() {
		return os.Remove(this.path)
	}
	return os.RemoveAll(this.path)
}

// 取得Writer
func (this *File) Writer() (*Writer, error) {
	writer := &Writer{}
	fp, err := os.OpenFile(this.path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	writer.file = fp
	writer.locker = &sync.Mutex{}
	return writer, nil
}

// 取得Appender
func (this *File) Appender() (*Appender, error) {
	appender := &Appender{}
	fp, err := os.OpenFile(this.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	appender.file = fp
	appender.locker = &sync.Mutex{}
	return appender, nil
}

// 取得Reader
func (this *File) Reader() (*Reader, error) {
	reader := &Reader{}
	fp, err := os.OpenFile(this.path, os.O_RDONLY, 0444)
	if err != nil {
		return nil, err
	}

	reader.file = fp
	return reader, nil
}
