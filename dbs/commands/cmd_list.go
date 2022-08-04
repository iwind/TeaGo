package commands

import (
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/utils/time"
	"os"
	"regexp"
	"strings"
	"time"
)

type ListModelsCommand struct {
	*cmd.Command
}

func (this *ListModelsCommand) Name() string {
	return "list models"
}

func (this *ListModelsCommand) Codes() []string {
	return []string{":db.list", ":db.ls", ":db.latest"}
}

func (this *ListModelsCommand) Usage() string {
	return ":db.[list|ls] [KEYWORD]\n   " + ":db.latest [KEYWORD]"
}

func (this *ListModelsCommand) Run() {
	keyword, _ := this.Arg(1)

	// 所有的模型
	db, err := dbs.Default()
	if err != nil {
		this.Error(err)
		return
	}

	config, err := db.Config()
	if err != nil {
		this.Error(err)
		return
	}

	pkg := config.Models.Package
	if len(pkg) == 0 {
		this.Println("'models.package' should be configured for db '" + db.Id() + "'")
		return
	}

	dir := files.NewFile(os.Getenv("GOPATH") + Tea.DS + pkg)
	if !dir.Exists() {
		this.Println("'" + pkg + "' does not exist")
		return
	}

	models := []string{}

	dir.Range(func(file *files.File) {
		if !file.IsFile() {
			return
		}

		if !strings.HasSuffix(file.Name(), ".go") {
			return
		}

		content, err := file.ReadAllString()
		if err != nil {
			this.Error(err)
			return
		}

		content = strings.Replace(content, "\n", " ", -1)
		content = strings.Replace(content, "\r", " ", -1)

		reg := regexp.MustCompile("type\\s+(\\w+)\\s+struct {.+}")
		if reg.MatchString(content) {
			match := reg.FindStringSubmatch(content)[1:]
			modelName := match[0]

			// 所有字段
			reg = regexp.MustCompile("(\\w+)\\s+(\\w+)\\s*`field:\"(\\w+)\"`")
			matches := reg.FindAllStringSubmatch(content, -1)
			if len(matches) == 0 {
				return
			}

			// 关键词
			if len(keyword) > 0 {
				reg := regexp.MustCompile("(?i)" + keyword)
				if !reg.MatchString(modelName) {
					return
				}
			}

			// 最近的...
			modifiedTime, _ := file.LastModified()
			if this.SubCode() == ":db.latest" {
				if time.Since(modifiedTime).Seconds() > 86400 {
					return
				}
			}

			if time.Since(modifiedTime).Seconds() < 3600 {
				modelName = logs.Sprintf("<code>"+modelName+"</code>   [modified at "+
					timeutil.Format("H:i:s")+"]", modifiedTime)
			}

			models = append(models, modelName)
		}
	})

	if len(models) > 0 {
		lists.Sort(models, func(i int, j int) bool {
			return models[i] < models[j]
		})
		for _, model := range models {
			this.Output(model + "\n")
		}
	} else {
		this.Output("can not find any models\n")
	}
}
