package files

import (
	"github.com/iwind/TeaGo/lists"
	"strings"
)

const (
	SortTypeModifiedTime        = SortType(1) // 按最后修改时间
	SortTypeModifiedTimeReverse = SortType(2) // 按最后修改时间倒排序
	SortTypeName                = SortType(3) // 按名称
	SortTypeNameReverse         = SortType(4) // 按名称倒排序
	SortTypeSize                = SortType(5) // 按文件尺寸
	SortTypeSizeReverse         = SortType(6) // 按文件尺寸倒排序
	SortTypeKind                = SortType(7) // 按文件类型
	SortTypeKindReverse         = SortType(8) // 按文件类型倒排序
)

type SortType int

// 对一组文件进行排序
func Sort(files []*File, sortType ... SortType) {
	realSortType := SortTypeName
	if len(sortType) > 0 {
		realSortType = sortType[0]
	}

	lists.Sort(files, func(i int, j int) bool {
		file1 := files[i]
		file2 := files[j]

		if realSortType == SortTypeName {
			return strings.ToLower(file1.Name()) < strings.ToLower(file2.Name())
		}

		if realSortType == SortTypeNameReverse {
			return strings.ToLower(file1.Name()) > strings.ToLower(file2.Name())
		}

		if realSortType == SortTypeModifiedTime || realSortType == SortTypeModifiedTimeReverse {
			time1, err := file1.LastModified()
			if err != nil {
				return true
			}

			time2, err := file2.LastModified()
			if err != nil {
				return true
			}

			if realSortType == SortTypeModifiedTime {
				return time1.UnixNano() < time2.UnixNano()
			}
			return time1.UnixNano() > time2.UnixNano()
		}

		if realSortType == SortTypeSize || realSortType == SortTypeSizeReverse {
			size1, err := file1.Size()
			if err != nil {
				return true
			}

			size2, err := file2.Size()
			if err != nil {
				return true
			}

			if realSortType == SortTypeSize {
				return size1 < size2
			}
			return size1 > size2
		}

		if realSortType == SortTypeKind || realSortType == SortTypeKindReverse {
			if file1.IsDir() {
				if file2.IsDir() {
					return strings.ToLower(file1.Name()) < strings.ToLower(file2.Name())
				}

				return realSortType == SortTypeKind
			}

			if file2.IsDir() {
				return realSortType == SortTypeKindReverse
			}

			return strings.ToLower(file1.Name()) < strings.ToLower(file2.Name())
		}

		return true
	})
}
