package dbs

import "strings"

type TableIndex struct {
	IsUnique    bool
	Name        string
	ColumnNames []string
	Type        string
	Comment     string
}

func (index *TableIndex) Definition() string {
	//示例：KEY `adId` (`adId`) USING BTREE COMMENT '广告ID'
	pieces := []string{"KEY", "`" + index.Name + "`", "(`" + strings.Join(index.ColumnNames, "`,`") + "`)", "USING " + index.Type}
	if len(index.Comment) > 0 {
		pieces = append(pieces, "COMMENT '"+index.Comment+"'")
	}
	return strings.Join(pieces, " ")
}
