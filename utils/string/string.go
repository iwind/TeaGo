package stringutil

import (
	"regexp"
	"crypto/md5"
	"fmt"
	"time"
	"math/rand"
	"strings"
	"strconv"
	"errors"
	"sync"
)

var reuseRegexpMap = map[string]*regexp.Regexp{}
var reuseRegexpMutex = &sync.Mutex{}

// 判断slice中是否包含某个字符串
func Contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

// 生成可重用的正则
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

// 计算字符串的md5
func Md5(source string) string {
	hash := md5.New()
	hash.Write([]byte(source))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// 取得随机字符串
// 代码来自 https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func Rand(n int) string {
	const randomLetterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	const (
		randomLetterIdxBits = 6                          // 6 bits to represent a letter index
		randomLetterIdxMask = 1<<randomLetterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		randomLetterIdxMax  = 63 / randomLetterIdxBits   // # of letter indices fitting in 63 bits
	)

	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), randomLetterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), randomLetterIdxMax
		}
		if idx := int(cache & randomLetterIdxMask); idx < len(randomLetterBytes) {
			b[i] = randomLetterBytes[idx]
			i--
		}
		cache >>= randomLetterIdxBits
		remain--
	}

	return string(b)
}

// 转换数字ID到字符串
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

// 翻转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ReplaceCommentsInJSON(jsonBytes []byte) []byte {
	commentReg, err := RegexpCompile("/([*]+((.|\n|\r)+?)[*]+/)|(\n\\s+//.+)")
	if err != nil {
		panic(err)
	}
	return commentReg.ReplaceAll(jsonBytes, []byte{})
}

// 从字符串中分析尺寸
func ParseFileSize(sizeString string) (float64, error) {
	if len(sizeString) == 0 {
		return 0, nil
	}

	reg, _ := RegexpCompile("^([\\d.]+)\\s*(b|byte|bytes|k|m|g|kb|mb|gb)$")
	matches := reg.FindStringSubmatch(strings.ToLower(sizeString))
	if len(matches) == 3 {
		size, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, err
		} else {
			unit := matches[2]
			if unit == "k" || unit == "kb" {
				size = size * 1024
			} else if unit == "m" || unit == "mb" {
				size = size * 1024 * 1024
			} else if unit == "g" || unit == "gb" {
				size = size * 1024 * 1024 * 1024
			}

			return size, nil
		}
	}
	return 0, errors.New("invalid string:" + sizeString)
}

// 对比版本号，返回-1，0，1三个值
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

	for i := 0; i < count1; i ++ {
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
