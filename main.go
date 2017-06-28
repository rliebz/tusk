package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/task"
	"gitlab.com/rliebz/tusk/ui"
)

func createCLIApp(tasks map[string]*task.Task) *cli.App {
	app := cli.NewApp()
	app.Name = "tusk"
	app.HelpName = "tusk"
	app.Usage = "a task runner built with simple configuration in mind"

	taskMap := make(map[string]*task.Task)

	// Create commands
	for name, task := range tasks {
		taskMap[name] = task
		app.Commands = append(app.Commands, createCommand(name, task))
	}

	// Update pretasks
	for _, task := range tasks {
		for _, name := range task.PreName {
			task.PreTasks = append(task.PreTasks, taskMap[name])
		}
	}

	return app
}

func createCommand(name string, task *task.Task) cli.Command {

	command := cli.Command{
		Name:  name,
		Usage: task.Usage,
		Action: func(c *cli.Context) error {
			return task.Execute()
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
	tasks, err := config.ReadTuskfile("tusk.yml")
	if err != nil {
		log.Fatal("Could not parse Tuskfile")
	}

	app := createCLIApp(tasks)
	if err := app.Run(os.Args); err != nil {
		ui.PrintError(err)
	}
}
