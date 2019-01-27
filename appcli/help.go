package appcli

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/rliebz/tusk/config/task"
	"github.com/urfave/cli"
)

// init sets the help templates for urfave/cli.
// nolint: lll, gochecknoinits
func init() {

	cli.HelpPrinter = helpPrinter
	cli.FlagNamePrefixer = flagPrefixer

	cli.AppHelpTemplate = `{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} <task> [task options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

Version:
   {{.Version}}{{end}}{{end}}{{if .Description}}

Description:
{{indent 3 .Description}}{{end}}{{if .VisibleCommands}}

Tasks:{{range .VisibleCategories}}{{$categoryName := .Name}}{{if $categoryName}}
   {{$categoryName}}:{{end}}{{range .VisibleCommands}}
   {{if $categoryName}}  {{end}}{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

Global Options:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

Copyright:
   {{.Copyright}}{{end}}
`
}

// ShowAppHelp shows the help for a given app.
func ShowAppHelp(app *cli.App) {
	app.Setup()
	cli.HelpPrinter(app.Writer, cli.AppHelpTemplate, app)
}

// helpPrinter includes the custom indent template function.
func helpPrinter(out io.Writer, templ string, data interface{}) {
	customFunc := map[string]interface{}{
		"indent": func(spaces int, text string) string {
			padding := strings.Repeat(" ", spaces)
			return padding + strings.Replace(text, "\n", "\n"+padding, -1)
		},
	}

	cli.HelpPrinterCustom(out, templ, data, customFunc)
}

// flagPrefixer formats the command-line flag usage.
func flagPrefixer(fullName, placeholder string) string {
	var output string

	parts := strings.Split(fullName, ",")
	for _, flagName := range parts {
		flagName = strings.Trim(flagName, " ")
		output = joinFlagString(output, flagName)
	}

	if strings.HasPrefix(output, "--") {
		output = "    " + output
	}

	if placeholder != "" {
		output = output + " <" + placeholder + ">"
	}

	return output
}

func joinFlagString(existing, flagName string) string {
	if existing == "" {
		return prependHyphens(flagName)
	}

	if len(flagName) == 1 {
		return prependHyphens(flagName) + ", " + existing
	}

	return existing + ", " + prependHyphens(flagName)
}

func prependHyphens(flagName string) string {
	if len(flagName) == 1 {
		return "-" + flagName
	}

	return "--" + flagName
}

func createCommandHelp(t *task.Task) string {
	// nolint: lll
	return fmt.Sprintf(`{{.HelpName}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{end}}{{if .Category}}

Category:
   {{.Category}}{{end}}{{if .Description}}

Description:
{{indent 3 .Description}}{{end}}%s{{if .VisibleFlags}}

Options:
   {{range  $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`, createArgsSection(t))
}

func createArgsSection(t *task.Task) string {
	argsTpl := `{{if .}}

Arguments:
   {{range  $index, $arg := .}}{{if $index}}
   {{end}}{{$arg}}{{end}}{{end}}`

	tpl := template.New(fmt.Sprintf("%s help", t.Name))
	tpl = template.Must(tpl.Parse(argsTpl))

	padArg := getArgPadder(t)

	argList := make([]string, 0, len(t.OrderedArgNames))
	for _, name := range t.OrderedArgNames {
		arg := t.Args[name]
		argText := fmt.Sprintf("%s%s", padArg(arg.Name), arg.Usage)
		argList = append(argList, strings.Trim(argText, " "))
	}

	var argsSection bytes.Buffer
	if err := tpl.Execute(&argsSection, argList); err != nil {
		panic(err)
	}

	return argsSection.String()
}

func getArgPadder(t *task.Task) func(string) string {
	maxLength := 0
	for _, arg := range t.OrderedArgNames {
		if len(arg) > maxLength {
			maxLength = len(arg)
		}
	}
	s := fmt.Sprintf("%%-%ds", maxLength+2)
	return func(text string) string {
		return fmt.Sprintf(s, text)
	}
}
