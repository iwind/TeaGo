package dbs

import "strings"

type TableIndex struct {
	IsUnique    bool
	Name        string
	ColumnNames []string
	Type        string
	Comment     string
}

func (this *TableIndex) Definition() string {
	//示例：KEY `adId` (`adId`) USING BTREE COMMENT '广告ID'
	var pieces = []string{}
	if this.IsUnique {
		pieces = append(pieces, "UNIQUE")
	}
	pieces = append(pieces, "KEY", "`"+this.Name+"`", "(`"+strings.Join(this.ColumnNames, "`,`")+"`)", "USING "+this.Type)
	if len(this.Comment) > 0 {
		pieces = append(pieces, "COMMENT '"+this.Comment+"'")
	}
	return strings.Join(pieces, " ")
}
