package stringutil

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iwind/TeaGo/rands"
	"hash"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var reuseRegexpMap = map[string]*regexp.Regexp{}
var reuseRegexpMutex = &sync.RWMutex{}

// Contains 判断slice中是否包含某个字符串
func Contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

// RegexpCompile 生成可重用的正则
func RegexpCompile(pattern string) (*regexp.Regexp, error) {
	reuseRegexpMutex.Lock()
	defer reuseRegexpMutex.Unlock()

	reg, ok := reuseRegexpMap[pattern]
	if ok {
		return reg, nil
	}

	reg, err := regexp.Compile(pattern)
	if err == nil {
		reuseRegexpMap[pattern] = reg
	}
	return reg, err
}

// Md5 Pool
var md5Pool = &sync.Pool{
	New: func() any {
		return md5.New()
	},
}

// Md5 计算字符串的md5
func Md5(source string) string {
	var m = md5Pool.Get().(hash.Hash)
	m.Write([]byte(source))
	var result = hex.EncodeToString(m.Sum(nil))
	m.Reset()
	md5Pool.Put(m)
	return result
}

// Rand 取得随机字符串
func Rand(n int) string {
	return rands.String(n)
}

// ConvertID 转换数字ID到字符串
func ConvertID(intId int64) string {
	const mapping = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code := ""
	size := int64(len(mapping))
	for intId >= size {
		mod := intId % size
		intId = intId / size

		code += mapping[mod : mod+1]
	}
	code += mapping[intId : intId+1]
	code = Reverse(code)

	return code
}

// Reverse 翻转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ParseFileSize 从字符串中分析尺寸
func ParseFileSize(sizeString string) (float64, error) {
	if len(sizeString) == 0 {
		return 0, nil
	}

	reg, _ := RegexpCompile("^([\\d.]+)\\s*(b|byte|bytes|k|m|g|t|p|e|z|y|kb|mb|gb|tb|pb|eb|zb|yb|)$")
	matches := reg.FindStringSubmatch(strings.ToLower(sizeString))
	if len(matches) == 3 {
		size, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, err
		} else {
			unit := matches[2]
			if unit == "k" || unit == "kb" {
				size = size * math.Pow(1024, 1)
			} else if unit == "m" || unit == "mb" {
				size = size * math.Pow(1024, 2)
			} else if unit == "g" || unit == "gb" {
				size = size * math.Pow(1024, 3)
			} else if unit == "t" || unit == "tb" {
				size = size * math.Pow(1024, 4)
			} else if unit == "p" || unit == "pb" {
				size = size * math.Pow(1024, 5)
			} else if unit == "e" || unit == "eb" {
				size = size * math.Pow(1024, 6)
			} else if unit == "z" || unit == "zb" {
				size = size * math.Pow(1024, 7)
			} else if unit == "y" || unit == "yb" {
				size = size * math.Pow(1024, 8)
			}

			return size, nil
		}
	}
	return 0, errors.New("invalid string:" + sizeString)
}

// VersionCompare 对比版本号，返回-1，0，1三个值
func VersionCompare(version1 string, version2 string) int8 {
	if len(version1) == 0 {
		if len(version2) == 0 {
			return 0
		}

		return -1
	}

	if len(version2) == 0 {
		return 1
	}

	pieces1 := strings.Split(version1, ".")
	pieces2 := strings.Split(version2, ".")
	count1 := len(pieces1)
	count2 := len(pieces2)

	for i := 0; i < count1; i++ {
		if i > count2-1 {
			return 1
		}

		piece1 := pieces1[i]
		piece2 := pieces2[i]
		len1 := len(piece1)
		len2 := len(piece2)

		if len1 == 0 {
			if len2 == 0 {
				continue
			}
		}

		maxLength := 0
		if len1 > len2 {
			maxLength = len1
		} else {
			maxLength = len2
		}

		piece1 = fmt.Sprintf("%0"+strconv.Itoa(maxLength)+"s", piece1)
		piece2 = fmt.Sprintf("%0"+strconv.Itoa(maxLength)+"s", piece2)

		if piece1 > piece2 {
			return 1
		}

		if piece1 < piece2 {
			return -1
		}
	}

	if count1 > count2 {
		return 1
	}

	if count1 == count2 {
		return 0
	}

	return -1
}

// JSONEncode JSON Encode
func JSONEncode(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

// JSONEncodePretty JSON Encode Pretty
func JSONEncodePretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "null"
	}
	return string(b)
}
