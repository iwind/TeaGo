package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"os"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/files"
	"strings"
	"regexp"
	"github.com/iwind/TeaGo/dbs"
)

type CheckModelCommand struct {
	*cmd.Command
}

func (command *CheckModelCommand) Name() string {
	return "check model's modification"
}

func (command *CheckModelCommand) Codes() []string {
	return []string{":db.check"}
}

func (command *CheckModelCommand) Usage() string {
	return ":db.check"
}

func (command *CheckModelCommand) Run() {
	// 所有的模型
	db, err := dbs.Default()
	if err != nil {
		command.Error(err)
		return
	}

	config, err := db.Config()
	if err != nil {
		command.Error(err)
		return
	}

	pkg := config.Models.Package
	if len(pkg) == 0 {
		command.Println("'models.package' should be configured for db '" + db.Id() + "'")
		return
	}

	dir := files.NewFile(os.Getenv("GOPATH") + Tea.DS + pkg)
	if !dir.Exists() {
		command.Println("'" + pkg + "' does not exist")
		return
	}

	command.Output("<code>checking ...</code>\n~~~\n")

	tables := []*dbs.Table{}                     // Model name => *Table
	models := map[string]map[string]*dbs.Field{} // Model name => { fields:... }
	modelFiles := map[string]string{}            // Model name => File name
	countIssues := 0

	dir.Range(func(file *files.File) {
		if !file.IsFile() {
			return
		}

		if !strings.HasSuffix(file.Name(), ".go") {
			return
		}

		content, err := file.ReadAllString()
		if err != nil {
			command.Error(err)
			return
		}

		content = strings.Replace(content, "\n", " ", -1)
		content = strings.Replace(content, "\r", " ", -1)

		// DAO
		reg := regexp.MustCompile("dbs.DAOObject{\\s*DB:.+,\\s*Table:\\s*\"(\\w+)\",\\s*Model:\\s*new\\((\\w+)\\),\\s*PkName:\\s*\"(\\w+)\"")
		if reg.MatchString(content) {
			match := reg.FindStringSubmatch(content)[1:]
			tableName := match[0]
			modelName := match[1]

			// 表信息
			table, err := db.FindTable(tableName)
			if err != nil || table == nil {
				path, _ := file.AbsPath()
				command.Output("<code>-[" + modelName + "] remove model</code>\n")
				command.outputFile(path)

				countIssues ++
				return
			}

			table.MappingName = modelName
			tables = append(tables, table)
		}

		reg = regexp.MustCompile("type\\s+(\\w+)\\s+struct {.+}")
		if reg.MatchString(content) {
			match := reg.FindStringSubmatch(content)[1:]
			modelName := match[0]

			modelFiles[modelName], _ = file.AbsPath()

			// 所有字段
			reg = regexp.MustCompile("(\\w+)\\s+(\\w+)\\s*`field:\"(\\w+)\"`")
			matches := reg.FindAllStringSubmatch(content, -1)
			fields := map[string]*dbs.Field{}
			for _, match := range matches {
				mappingName := match[1]
				dataTypeString := match[2]
				fieldName := match[3]

				field := new(dbs.Field)
				field.Name = fieldName
				field.MappingName = mappingName
				field.MappingKindName = dataTypeString

				fields[field.Name] = field
			}

			models[modelName] = fields
		}
	})

	// 检查现有table
	for _, table := range tables {
		modelName := table.MappingName
		oldFields, found := models[modelName]
		if !found {
			command.Output("+[" + modelName + "] gen model\n")
		} else {
			// 新增字段或修改字段
			for _, field := range table.Fields {
				oldField, found := oldFields[field.Name]
				if !found {
					command.Output("<code>+[" + modelName + "] field: " + command.convertFieldNameStyle(field.Name) + " " + field.ValueTypeName() + " `field:\"" + field.Name + "\"` // " + field.Comment + "</code>\n")
					command.outputFile(modelFiles[modelName])
					countIssues ++
				} else {
					// 对比
					if field.ValueTypeName() != oldField.MappingKindName {
						command.Output("<code>*[" + modelName + "] field: " + oldField.MappingName + " " + field.ValueTypeName() + " `field:\"" + field.Name + "\"` // " + field.Comment + "</code>\n")
						command.outputFile(modelFiles[modelName])
						countIssues ++
					}
				}
			}

			// 删除字段
			for _, oldField := range oldFields {
				field := table.FindFieldWithName(oldField.Name)
				if field == nil {
					command.Output("<code>-[" + modelName + "] field: " + oldField.MappingName + "</code>\n")
					command.outputFile(modelFiles[modelName])
					countIssues ++
				}
			}
		}
	}

	if countIssues == 0 {
		command.Output("<ok>Everything goes ok</ok>\n")
	} else {
		command.Output("~~~\n")
		command.Output("<error>There are", countIssues, "issues to be fixed</error>\n")
	}
}

func (command *CheckModelCommand) convertFieldNameStyle(fieldName string) string {
	pieces := strings.Split(fieldName, "_")
	newPieces := []string{}
	for _, piece := range pieces {
		newPieces = append(newPieces, strings.ToUpper(string(piece[0]))+string(piece[1:]))
	}
	return strings.Join(newPieces, "")
}

func (command *CheckModelCommand) outputFile(file string) {
	goPath := os.Getenv("GOPATH")
	command.Output("   ", strings.TrimPrefix(file, goPath), "\n")
}
