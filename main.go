package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"

	"gitlab.com/rliebz/tusk/script"
	"gitlab.com/rliebz/tusk/ui"
)

// Task is a single task to be run by CLI
type Task struct {
	Args   []Arg    `yaml:",omitempty"`
	Pre    []string `yaml:",omitempty"`
	Script []script.Script
	Usage  string
}

// Arg represents a command line argument
type Arg struct {
	Name        string
	Alias       []string // TODO: How does urfave/cli support?
	Default     string
	Environment string
	Usage       string
}

func run(task Task) error {
	for _, script := range task.Script {
		if err := script.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func createCLIApp(tasks map[string]Task) *cli.App {
	app := cli.NewApp()
	app.Name = "tusk"
	app.HelpName = "tusk"
	app.Usage = "a task runner built with simple configuration in mind"

	for name, task := range tasks {
		app.Commands = append(app.Commands, createCommand(name, task))
	}
	return app
}

func readTuskfile(filename string) (map[string]Task, error) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	tasks := make(map[string]Task)
	err = yaml.Unmarshal(data, &tasks)
	if err != nil {
		log.Printf("error: %v\n", err)
		return nil, err
	}

	return tasks, nil
}

func createCommand(name string, task Task) cli.Command {

	command := cli.Command{
		Name:  name,
		Usage: task.Usage,
		Action: func(c *cli.Context) error {
			return run(task)
		},
	}

	for _, arg := range task.Args {
		// TODO: Flag types
		flag := cli.StringFlag{
			Name:   arg.Name,
			Value:  arg.Default,
			Usage:  arg.Usage,
			EnvVar: arg.Environment,
		}
		command.Flags = append(command.Flags, flag)
	}
	return command
}

func main() {
	tasks, err := readTuskfile("tusk.yml")
	if err != nil {
		log.Fatal("Could not parse Tuskfile")
	}

	app := createCLIApp(tasks)
	if err := app.Run(os.Args); err != nil {
		ui.PrintError(err)
	}
}
