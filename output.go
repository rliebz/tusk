package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

var (
	blue   = color.New(color.FgBlue).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

func printCommand(cmd *exec.Cmd) {
	message := fmt.Sprintf(
		"[%s] %s\n",
		blue("Running"),
		yellow(strings.Join(cmd.Args, " ")),
	)
	fmt.Printf(message)
}

func printCommandStdout(text string) {
	message := fmt.Sprintf("  %s %s\n", green("=>"), text)
	fmt.Printf(message)
}

func printCommandStderr(text string) {
	message := fmt.Sprintf("  %s %s\n", yellow("=>"), text)
	fmt.Printf(message)
}

func printError(err error) {
	if err != nil {
		message := fmt.Sprintf("[%s] %s\n", red("ERROR"), err.Error())
		os.Stderr.WriteString(message)
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}
