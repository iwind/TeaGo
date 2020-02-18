package commands

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/utils/string"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GenModelCommand struct {
	*cmd.Command
}

func (this *GenModelCommand) Name() string {
	return "generate model and dao files"
}

func (this *GenModelCommand) Codes() []string {
	return []string{":db.gen"}
}

func (this *GenModelCommand) Usage() string {
	return ":db.gen [MODEL_NAME] [-db=[DB ID] -dir=[TARGET DIR]]"
}

func (this *GenModelCommand) Run() {
	model, found := this.Arg(1)
	if !found {
		this.Error(errors.New("please specify model name"))
		return
	}
	dbId, found := this.Param("db")
	var db *dbs.DB
	var err error
	if found {
		db, err = dbs.Instance(dbId)
		if err != nil {
			this.Error(err)
			return
		}
	} else {
		db, err = dbs.Default()
		if err != nil {
			this.Error(err)
			return
		}
	}

	// 模型目录
	subPackage := "models"
	dir, _ := this.Param("dir")
	config, _ := db.Config()
	if len(config.Models.Package) > 0 {
		dir = strings.TrimSuffix(config.Models.Package+Tea.DS+dir, Tea.DS)
	}

	packagePieces := strings.Split(model, ".")
	if len(packagePieces) > 0 {
		model = packagePieces[len(packagePieces)-1]

		if len(dir) > 0 {
			dir += Tea.DS + strings.Join(packagePieces[:len(packagePieces)-1], Tea.DS)

			dirFile := files.NewFile(dir)
			if !dirFile.Exists() {
				dirFile.MkdirAll()
			}
		} else {
			subPackage = packagePieces[len(packagePieces)-2]
		}
	}

	if len(dir) > 0 {
		subPackage = filepath.Base(dir)
	}

	// 取得对应表
	subTableName, err := this.modelToTable(model)
	if err != nil {
		this.Error(err)
		return
	}
	tableName := db.TablePrefix() + subTableName

	tableNames, err := db.TableNames()
	if err != nil {
		this.Error(err)
		return
	}
	lowerTableName := strings.Replace(strings.ToLower(tableName), "_", "", -1)
	for _, dbTableName := range tableNames {
		if strings.ToLower(strings.Replace(dbTableName, "_", "", -1)) == lowerTableName {
			tableName = dbTableName
			break
		}
	}

	table, err := db.FindTable(tableName)
	if err != nil {
		this.Error(err)
		return
	}
	if table == nil {
		this.Println("not found table named '" + tableName + "'")
		return
	}

	// Model
	modelString := `package ` + subPackage + `

// ` + strings.Replace(table.Comment, "\n", " ", -1) + `
type ` + model + ` struct {`
	modelString += "\n"
	var primaryKey = ""
	var primaryKeyType = ""
	fieldNames := []string{}
	for _, field := range table.Fields {
		fieldNames = append(fieldNames, field.Name)

		if field.IsPrimaryKey {
			primaryKeyType = field.ValueTypeName()
			primaryKey = field.Name
		}
	}

	for _, field := range table.Fields {
		var attr = this.convertFieldNameStyle(field.Name)
		var dataType = field.ValueTypeName()
		modelString += "\t" + attr + " " + dataType + " `field:\"" + field.Name + "\"` //" + field.Comment + "\n"
	}

	modelString += "}"
	modelString += "\n\n"

	// Operator
	modelString += `type ` + model + `Operator struct {`
	modelString += "\n"

	for _, field := range table.Fields {
		var attr = this.convertFieldNameStyle(field.Name)
		modelString += "\t" + attr + " interface{}" + " // " + field.Comment + "\n"
	}

	modelString += "}"
	modelString += "\n"
	modelString += `
func New` + model + `Operator() *` + model + `Operator {
	return &` + model + `Operator{}
}
`

	formatted, err := format.Source([]byte(modelString))
	if err == nil {
		modelString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Println("Model:")
		fmt.Println("~~~")
		fmt.Println(modelString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_model.go"
		file := files.NewFile(target)
		if file.Exists() {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(modelString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}

	// DAO
	daoString := `package ` + subPackage + `

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/Tea"
)
`

	if stringutil.Contains(fieldNames, "state") {
		daoString += `const (
	${model}StateEnabled = 1 // 已启用
	${model}StateDisabled = 0 // 已禁用
)
`
	}

	daoString += `type ` + model + `DAO dbs.DAO

func New` + model + `DAO() *` + model + `DAO {
	return dbs.NewDAO(&` + model + `DAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "` + tableName + `",
			Model:  new(` + model + `),
			PkName: "` + primaryKey + `",
		},
	}).(*` + model + `DAO)
}

var Shared` + model + `DAO = New` + model + `DAO()` + "\n"

	// state
	if stringutil.Contains(fieldNames, "state") {
		daoString += `
// 启用条目
func (this *${daoName}) Enable${model}(${pkName} ${pkNameType}) (rowsAffected int64, err error) {
	return this.Query().
		Pk(${pkName}).
		Set("state", ${model}StateEnabled).
		Update()
}

// 禁用条目
func (this *${daoName}) Disable${model}(${pkName} ${pkNameType}) (rowsAffected int64, err error) {
	return this.Query().
		Pk(${pkName}).
		Set("state", ${model}StateDisabled).
		Update()
}

// 查找启用中的条目
func (this *${daoName}) FindEnabled${model}(${pkName} ${pkNameType}) (*${model}, error) {
	result, err := this.Query().
		Pk(${pkName}).
		Attr("state", ${model}StateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*${model}), err
}
`
	}

	if stringutil.Contains(fieldNames, "name") && table.FindFieldWithName("name").ValueTypeName() == "string" {
		daoString += `// 根据主键查找名称
func (this *${daoName}) Find${model}Name(${pkName} ${pkNameType}) (string, error) {
	name, err := this.Query().
		Pk(${pkName}).
		Result("name").
		FindCol("")
	return name.(string), err
}
`
	}

	daoString = strings.Replace(daoString, "${daoName}", model+"DAO", -1)
	daoString = strings.Replace(daoString, "${pkName}", primaryKey, -1)
	daoString = strings.Replace(daoString, "${pkNameType}", primaryKeyType, -1)
	daoString = strings.Replace(daoString, "${model}", model, -1)

	formatted, err = format.Source([]byte(daoString))
	if err == nil {
		daoString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Print("\n\n")
		fmt.Println("DAO:")
		fmt.Println("~~~")
		fmt.Println(daoString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_dao.go"
		file := files.NewFile(target)
		if file.Exists() {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(daoString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}

	// test
	testString := `package ` + subPackage + `
import (
	_ "github.com/go-sql-driver/mysql"
)
`
	formatted, err = format.Source([]byte(testString))
	if err == nil {
		testString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Print("\n\n")
		fmt.Println("DAO Test:")
		fmt.Println("~~~")
		fmt.Println(testString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_dao_test.go"
		file := files.NewFile(target)
		if file.Exists() {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(testString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}
}

func (this *GenModelCommand) modelToTable(modelName string) (string, error) {
	var tableName = modelName + "s"

	// ies
	reg, err := stringutil.RegexpCompile("(?i)(cit|categor|part|activit|stor|famil|bab|lad|librar|difficult|histor|compan|deliver|cop|stud|enem|repl|glor|communit|propert)ys")
	if err != nil {
		return tableName, err
	}
	tableName = reg.ReplaceAllString(tableName, "${1}ies")

	// oes
	reg, err = stringutil.RegexpCompile("(?i)(hero|potato|tomato|echo|tornado|torpedo|domino|veto|mosquito|negro|mango|buffalo|volcano|match|dish|brush|branch|dress|glass|bus|class|boss|process|box|fox|watch|index)s")
	if err != nil {
		return tableName, err
	}
	tableName = reg.ReplaceAllString(tableName, "${1}es")

	// ves
	for find, replace := range map[string]string{
		"leafs$":          "leaves",
		"halfs$":          "halves",
		"wolfs$":          "wolves",
		"shiefs$":         "shieves",
		"shelfs$":         "shelves",
		"knifes$":         "knives",
		"wifes$":          "wives",
		"(goods|money)s$": "$1",
	} {
		reg, err = stringutil.RegexpCompile("(?i)" + find)
		if err != nil {
			return tableName, err
		}
		tableName = reg.ReplaceAllString(tableName, replace)
	}

	return tableName, nil
}

func (this *GenModelCommand) convertFieldNameStyle(fieldName string) string {
	pieces := strings.Split(fieldName, "_")
	newPieces := []string{}
	for _, piece := range pieces {
		newPieces = append(newPieces, strings.ToUpper(string(piece[0]))+string(piece[1:]))
	}
	return strings.Join(newPieces, "")
}

func (this *GenModelCommand) convertToUnderlineName(modelName string) string {
	reg := regexp.MustCompile("[A-Z]")
	return strings.TrimPrefix(reg.ReplaceAllStringFunc(modelName, func(s string) string {
		return "_" + strings.ToLower(s)
	}), "_")
}
