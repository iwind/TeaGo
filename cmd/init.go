package cmd

func init() {
	Register(new(HelpCommand))
	Register(new(QuitCommand))
}
