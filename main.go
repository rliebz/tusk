package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

// TODO: Handle errors
func execCommand(command string) error {
	parts := strings.Fields(command)
	head := parts[0]
	args := parts[1:]

	cmd := exec.Command(head, args...)
	printCommand(cmd)

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	// TODO: Is it possible to keep the output ordered and separate?
	cmd.Stdout = pw
	cmd.Stderr = pw

	scanner := bufio.NewScanner(pr)
	go func() {
		for scanner.Scan() {
			printCommandStdout(scanner.Text())
		}
	}()

	// stdoutReader, err := cmd.StdoutPipe()
	// if err != nil {
	// 	printError(err)
	// 	return err
	// }
	// stdoutScanner := bufio.NewScanner(stdoutReader)
	// go func() {
	// 	for stdoutScanner.Scan() {
	// 		printCommandStdout(stdoutScanner.Text())
	// 	}
	// }()

	// // TODO: Fix race condition for stderr/stdout ordering
	// stderrReader, err := cmd.StderrPipe()
	// if err != nil {
	// 	printError(err)
	// 	return err
	// }
	// stderrScanner := bufio.NewScanner(stderrReader)
	// go func() {
	// 	for stderrScanner.Scan() {
	// 		printCommandStderr(stderrScanner.Text())
	// 	}
	// }()

	if err := cmd.Run(); err != nil {
		printError(err)
		return err
	}

	// TODO: Do we need these?
	pr.Close()
	pw.Close()

	return nil
}

func testCommand(test string) error {
	args := strings.Fields(test)
	_, err := exec.Command("test", args...).Output()
	return err
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

// TODO: Check for errors
func run(task Task) error {
	for _, script := range task.Script {
		runScript(script)
	}
	return nil
}

// TODO: Check for errors
func runScript(script Script) {

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
		if err := testCommand(test); err != nil {
			fmt.Printf("Test failed: %s\n", test)
			return
		}
	}

	for _, command := range script.Run {
		// TODO: Capture return value
		execCommand(command)
	}
}

func createCLIApp() {
	app := cli.NewApp()
	app.Name = "tusk"
	app.HelpName = "tusk"
	app.Usage = "a task runner built with simple configuration in mind"

	tasks, err := readTuskfile("tusk.yml")
	if os.IsNotExist(err) {
		fmt.Printf("No tusk.yml found\n\n")
	}

	for name, task := range tasks {
		app.Commands = append(app.Commands, createCommand(name, task))
	}

	app.Run(os.Args)
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
	createCLIApp()
}
