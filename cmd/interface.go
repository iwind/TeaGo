package cmd

type CommandInterface interface {
	Name() string
	Usage() string
	Codes() []string
	SubCode() string
	Run()
}
