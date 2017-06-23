package task

// Arg represents a command line argument
type Arg struct {
	Name        string
	Alias       []string // TODO: How does urfave/cli support?
	Default     string
	Environment string
	Usage       string
}
