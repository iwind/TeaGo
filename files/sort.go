package files

import (
	"github.com/iwind/TeaGo/lists"
)

const (
	SortTypeModifiedTime        = SortType(1) // 按最后修改时间
	SortTypeModifiedTimeReverse = SortType(2) // 按最后修改时间倒排序
	SortTypeName                = SortType(3) // 按名称
	SortTypeNameReverse         = SortType(4) // 按名称倒排序
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
			return file1.Name() < file2.Name()
		}

		if realSortType == SortTypeNameReverse {
			return file1.Name() > file2.Name()
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

		return true
	})
}
