package commands

import (
	"fmt"
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"regexp"
)

type CompareDBCommandOptions struct {
	Tables    bool
	Functions bool
}

type CompareDBIssue struct {
	dbId string
	sql  string
}

type CompareDBCommand struct {
	*cmd.Command
}

var compareIssues = map[int]*CompareDBIssue{}

func (this *CompareDBCommand) Name() string {
	return "compare two databases"
}

func (this *CompareDBCommand) Codes() []string {
	return []string{":db.compare"}
}

func (this *CompareDBCommand) Usage() string {
	return ":db.compare db1 db2"
}

func (this *CompareDBCommand) Run() {
	dbId1, found := this.Arg(1)
	if !found {
		this.Output("need db1, db2 to compare\n")
		return
	}

	dbId2, found := this.Arg(2)
	if !found {
		this.Output("need db2 to compare\n")
		return
	}

	db1, err := dbs.NewInstance(dbId1)
	if err != nil {
		this.Error(err)
		return
	}
	defer func() {
		_ = db1.Close()
	}()

	db2, err := dbs.NewInstance(dbId2)
	if err != nil {
		this.Error(err)
		return
	}
	defer func() {
		_ = db2.Close()
	}()

	options := &CompareDBCommandOptions{
		Tables:    true,
		Functions: true,
	}
	this.compareTables(db1, db2, options)
}

func (this *CompareDBCommand) compareTables(db1 *dbs.DB, db2 *dbs.DB, options *CompareDBCommandOptions) {
	version1, err := db1.FindCol(0, "SELECT VERSION()")
	if err != nil {
		this.Error(err)
		return
	}

	version2, err := db2.FindCol(0, "SELECT VERSION()")
	if err != nil {
		this.Error(err)
		return
	}
	isMySQL80 := stringutil.VersionCompare(types.String(version1), "8.0.0") >= 0 || stringutil.VersionCompare(types.String(version2), "8.0.0") >= 0

	compareIssues = map[int]*CompareDBIssue{}

	tableNames1, err := db1.TableNames()
	if err != nil {
		this.Error(err)
		return
	}

	tableNames2, err := db2.TableNames()
	if err != nil {
		this.Error(err)
		return
	}

	countIssues := 0

	this.Output("comparing ...\n")

	// 增加的或修改
	if options.Tables {
		this.Output("[tables]\n")
		for _, tableName1 := range tableNames1 {
			table1, _ := db1.FindFullTable(tableName1)
			table2, err := db2.FindFullTable(tableName1)

			// table2不存在
			if err != nil || table2 == nil {
				countIssues++
				this.Output("<code>+"+tableName1+" table</code>", fmt.Sprintf("[%d]", countIssues), "\n")
				if len(table1.Code) > 0 {
					reg := regexp.MustCompile(" AUTO_INCREMENT=\\d+")
					table1.Code = reg.ReplaceAllString(table1.Code, "")
					this.Output("   suggest: \n  " + table1.Code + ";\n")
					compareIssues[countIssues] = &CompareDBIssue{
						dbId: db2.Id(),
						sql:  table1.Code,
					}
				}
				continue
			}

			// 对比字段
			for _, field := range table1.Fields {
				field2 := table2.FindFieldWithName(field.Name)
				if field2 == nil {
					countIssues++
					this.Output("<code>+"+tableName1+" field: "+field.Name+" "+field.Definition()+"</code>", fmt.Sprintf("[%d]", countIssues), "\n")
					this.Output("   suggest: ALTER TABLE `" + tableName1 + "` ADD `" + field.Name + "` " + field.Definition() + ";\n")
					compareIssues[countIssues] = &CompareDBIssue{
						dbId: db2.Id(),
						sql:  "ALTER TABLE `" + tableName1 + "` ADD `" + field.Name + "` " + field.Definition(),
					}
				} else {
					if field.Definition() != field2.Definition() {
						// 检查是否是MySQL 8.0 以后的整型
						if isMySQL80 && (field.Type == "int" || field.Type == "tinyint" || field.Type == "bigint") {
							fullType1 := regexp.MustCompile(`\(\d+\)`).ReplaceAllString(field.FullType, "")
							fullType2 := regexp.MustCompile(`\(\d+\)`).ReplaceAllString(field2.FullType, "")
							if fullType1 == fullType2 {
								continue
							}
						}

						countIssues++
						this.Output("<code>*"+tableName1+" field: "+field.Name+" "+field.Definition()+
							"</code>", fmt.Sprintf("[%d]", countIssues), "\n   from "+field2.Name+" "+field2.Definition()+"\n")
						this.Output("   suggest: ALTER TABLE `" + tableName1 + "` MODIFY `" + field.Name + "` " + field.Definition() + ";\n")
						compareIssues[countIssues] = &CompareDBIssue{
							dbId: db2.Id(),
							sql:  "ALTER TABLE `" + tableName1 + "` MODIFY `" + field.Name + "` " + field.Definition(),
						}
					}
				}
			}

			for _, field := range table2.Fields {
				field1 := table1.FindFieldWithName(field.Name)
				if field1 == nil {
					countIssues++
					this.Output("<code>-"+tableName1+" field: "+field.Name+"</code>", fmt.Sprintf("[%d]", countIssues), "\n")
					this.Output("   suggest: ALTER TABLE `" + tableName1 + "` DROP COLUMN `" + field.Name + "`;\n")
					compareIssues[countIssues] = &CompareDBIssue{
						dbId: db2.Id(),
						sql:  "ALTER TABLE `" + tableName1 + "` DROP COLUMN `" + field.Name + "`",
					}
				}
			}

			// @TODO 对比选项

			// 对比partitions
			for _, partition := range table1.Partitions {
				partition2 := table2.FindPartitionWithName(partition.Name)
				if partition2 == nil {
					countIssues++
					this.Output("<code>+" + tableName1 + " partition: " + partition.Method + " (" + partition.Expression + ") " + partition.Name + " (" + partition.Description + ")</code>\n")
				} else {
					if partition.Method != partition2.Method ||
						partition.Description != partition2.Description ||
						partition.Expression != partition2.Expression {
						countIssues++
						this.Output("<code>*" + tableName1 + " partition: " + partition.Method + " (" + partition.Expression + ") " + partition.Name + " (" + partition.Description + ")</code>\n")
						this.Output("   from " + partition2.Method + " (" + partition2.Expression + ") " + partition2.Name + " (" + partition2.Description + ")\n")
					}
				}
			}

			for _, partition := range table2.Partitions {
				partition1 := table1.FindPartitionWithName(partition.Name)
				if partition1 == nil {
					countIssues++
					this.Output("<code>-" + tableName1 + " partition: " + partition.Method + " (" + partition.Expression + ") " + partition.Name + " (" + partition.Description + ")</code>\n")
				}
			}

			// 对比索引
			for _, index := range table1.Indexes {
				index2 := table2.FindIndexWithName(index.Name)
				if index2 == nil {
					countIssues++
					this.Output("<code>+"+tableName1+" index: "+index.Definition()+"</code>", fmt.Sprintf("[%d]", countIssues), "\n")
					this.Output("   suggest: ALTER TABLE `" + tableName1 + "` ADD " + index.Definition() + ";\n")
					compareIssues[countIssues] = &CompareDBIssue{
						dbId: db2.Id(),
						sql:  "ALTER TABLE `" + tableName1 + "` ADD " + index.Definition(),
					}
				} else {
					if index.Definition() != index2.Definition() {
						countIssues++
						this.Output("<code>*" + tableName1 + " index: " + index.Definition() + "</code>\n")
						this.Output("   from " + index2.Definition() + "\n")
					}
				}
			}

			for _, index := range table2.Indexes {
				index1 := table1.FindIndexWithName(index.Name)
				if index1 == nil {
					countIssues++
					this.Output("<code>-"+tableName1+" index: "+index.Definition()+"</code>", fmt.Sprintf("[%d]", countIssues), "\n")
					this.Output("   suggest: ALTER TABLE `" + tableName1 + "` DROP INDEX `" + index.Name + "`;\n")
					compareIssues[countIssues] = &CompareDBIssue{
						dbId: db2.Id(),
						sql:  "ALTER TABLE `" + tableName1 + "` DROP INDEX `" + index.Name + "`",
					}
				}
			}

			// @TODO 对比Triggers

		}

		// 减少的表
		for _, tableName2 := range tableNames2 {
			table1, err := db1.FindTable(tableName2)
			if err != nil || table1 == nil {
				countIssues++
				this.Output("<code>-"+tableName2+"</code>", fmt.Sprintf("[%d]", countIssues), "\n")
				this.Output("   suggest: DROP TABLE `" + tableName2 + "`;\n")
				compareIssues[countIssues] = &CompareDBIssue{
					dbId: db2.Id(),
					sql:  "DROP TABLE `" + tableName2 + "`",
				}
				continue
			}
		}
	}

	// @TODO 对比Views

	// 对比Functions
	if options.Functions {
		this.Output("\n[functions]\n")
		functions1, _ := db1.FindFunctions()
		functions2, _ := db2.FindFunctions()

		for _, function := range functions1 {
			function2 := this.findFunctionWithName(functions2, function.Name)
			if function2 == nil {
				countIssues++
				this.Output("<code>+" + function.Name + " function</code>\n")
			} else {
				// 去除definer后比较
				code := this.cleanFunctionCode(function.Code)
				code2 := this.cleanFunctionCode(function2.Code)
				if code != code2 {
					countIssues++
					this.Output("<code>*" + function.Name + " function</code>\n")
				}
			}
		}

		for _, function := range functions2 {
			function1 := this.findFunctionWithName(functions1, function.Name)
			if function1 == nil {
				countIssues++
				this.Output("<code>-" + function.Name + " function</code>\n")
			}
		}
	}

	// @TODO 对比Procedures

	// @TODO 对比Events

	if countIssues > 0 {
		this.Output("[result]\n")

		if len(compareIssues) > 0 {
			this.Output("<warn>There are", countIssues, "issues to be fixed, you can exec ':db.fix [ISSUE ID]' to fix some of them</warn>\n")
		} else {
			this.Output("<warn>There are", countIssues, "issues to be fixed</warn>\n")
		}
	} else {
		this.Output("<ok>Both database have the same schema</ok>\n")
	}
}

func (this *CompareDBCommand) findFunctionWithName(functions []*dbs.Function, name string) *dbs.Function {
	for _, function := range functions {
		if function.Name == name {
			return function
		}
	}
	return nil
}

func (this *CompareDBCommand) cleanFunctionCode(code string) string {
	definerReg := regexp.MustCompile("DEFINER=[^ ]+")
	code = definerReg.ReplaceAllString(code, "")

	optionsReg := regexp.MustCompile("(\\s*READS SQL DATA)|(\\s*DETERMINISTIC)")
	code = optionsReg.ReplaceAllString(code, "")
	return code
}
