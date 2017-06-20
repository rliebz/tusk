package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

// TODO: Read from file
var data = `
test:
  usage: run application tests
  args:
    - name: env
      alias:
        - environment
      default: local
      usage: An environment in which to run
      environment: TUSK_ENV
  pre:
    - bootstrap
  script:
    - run:
      - echo "Hello, world!"
    - run:
        # Only things for a development environment will run here
        - echo "Foo exists!"
      when:
        exists:
          - foo.json
        os:
          - darwin
        test:
          - -z "$RAILS_ENV"
          - -z "$RACK_ENV"

`

func execCommand(cmd string) {
	fmt.Printf("Running command: %v\n", cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	args := parts[1:len(parts)]

	out, err := exec.Command(head, args...).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	if len(out) != 0 {
		fmt.Printf("%s", out)
	}

}

// Task is a single task to be run by CLI
type Task struct {
	Args   []Arg    `yaml:",omitempty"`
	Pre    []string `yaml:",omitempty"`
	Script []Script
	Usage  string
}

// Arg is a command line argument
type Arg struct {
	Name        string
	Alias       []string // TODO: How does urfave/cli support?
	Default     string
	Environment string
	Usage       string
}

// Script is a single script within a task
type Script struct {
	When struct {
		Exists []string `yaml:",omitempty"`
		OS     []string `yaml:",omitempty"`
		Test   []string `yaml:",omitempty"`
	} `yaml:",omitempty"`
	Run []string
}

func run(task Task) error {
	// TODO: Check for errors
	for _, script := range task.Script {
		runScript(script)
	}
	return nil
}

func runScript(script Script) {
	// TODO: Check for errors

	for _, f := range script.When.Exists {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", f)
			return
		}
	}

	for _, os := range script.When.OS {
		if runtime.GOOS != os {
			fmt.Printf("Unexpected Architecture: %s\n", os)
			return
		}
	}

	for _, test := range script.When.Test {
		// TODO: Execute tests with `exec`
		fmt.Printf("Skipping test: %s\n", test)
	}

	for _, command := range script.Run {
		// TODO: Capture return value
		execCommand(command)
	}
}

func parseArgs(tasks map[string]Task) {
	app := cli.NewApp()
	app.Name = "tusk"
	app.HelpName = "tusk"
	app.Usage = "a task runner built for simple configuration"
	// app.UsageText = ""

	for name, task := range tasks {
		app.Commands = append(app.Commands, createCommand(name, task))
	}

	// sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
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

	tasks := make(map[string]Task)
	err := yaml.Unmarshal([]byte(data), &tasks)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	parseArgs(tasks)

}
