package commands

import (
	"github.com/iwind/TeaGo/cmd"
	"github.com/iwind/TeaGo/dbs"
	"github.com/go-yaml/yaml"
)

type InfoCommand struct {
	*cmd.Command
}

func (command *InfoCommand) Name() string {
	return "print database info"
}

func (command *InfoCommand) Codes() []string {
	return []string{":db.info"}
}

func (command *InfoCommand) Usage() string {
	return ":db.info"
}

func (command *InfoCommand) Run() {
	db, err := dbs.Default()
	if err != nil {
		command.Error(err)
		return
	}

	config, _ := db.Config()
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		command.Error(err)
		return
	}

	command.Output("<code>" + string(yamlBytes) + "</code>")
}
