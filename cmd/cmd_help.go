package cmd

type HelpCommand struct {
	*Command
}

func (this *HelpCommand) Name() string {
	return "Command help"
}

func (this *HelpCommand) Usage() string {
	return "help [COMMAND]"
}

func (this *HelpCommand) Codes() []string {
	return []string{"help"}
}

func (this *HelpCommand) Run() {
	commandCode, found := this.Arg(1)
	if !found {
		commandCode = ""
	}

	command, found := commandsPtrMap[commandCode]
	if !found {
		this.Output("command '"+commandCode+"' not found", "\n")
		return
	}

	this.Output(command.Name() + "\n")
	this.Output("Usage:\n   " + command.Usage() + "\n")
}
