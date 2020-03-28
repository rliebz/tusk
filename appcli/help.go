package appcli

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
	"github.com/urfave/cli"
)

// init sets the help templates for urfave/cli.
// nolint: lll, gochecknoinits
func init() {
	// These are both used, so both must be overridden
	cli.HelpPrinterCustom = wrapPrinter(cli.HelpPrinterCustom)
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		cli.HelpPrinterCustom(w, templ, data, nil)
	}

	cli.FlagNamePrefixer = flagPrefixer

	cli.AppHelpTemplate = `{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} <task> [task options]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

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
func ShowAppHelp(logger *ui.Logger, app *cli.App) {
	app.Setup()
	cli.HelpPrinter(logger.Stdout, cli.AppHelpTemplate, app)
}

type helpPrinterCustom = func(io.Writer, string, interface{}, map[string]interface{})

// helpPrinter includes the custom indent template function.
func wrapPrinter(p helpPrinterCustom) helpPrinterCustom {
	return func(w io.Writer, tpl string, data interface{}, funcs map[string]interface{}) {
		customFuncs := map[string]interface{}{
			"indent": func(spaces int, text string) string {
				padding := strings.Repeat(" ", spaces)
				return padding + strings.ReplaceAll(text, "\n", "\n"+padding)
			},
		}

		for k, v := range funcs {
			customFuncs[k] = v
		}

		p(w, tpl, data, customFuncs)
	}
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

func createCommandHelp(t *runner.Task) string {
	// nolint: lll
	return fmt.Sprintf(`{{.HelpName}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{end}}{{end}}{{if .Category}}

Category:
   {{.Category}}{{end}}{{if .Description}}

Description:
{{indent 3 .Description}}{{end}}%s{{if .VisibleFlags}}

Options:
   {{range  $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`, createArgsSection(t))
}

func createArgsSection(t *runner.Task) string {
	argsTpl := `{{if .}}

Arguments:
   {{range  $index, $arg := .}}{{if $index}}
   {{end}}{{$arg}}{{end}}{{end}}`

	tpl := template.New(fmt.Sprintf("%s help", t.Name))
	tpl = template.Must(tpl.Parse(argsTpl))

	padArg := getArgPadder(t)

	args := make([]string, 0, len(t.Args))
	for _, arg := range t.Args {
		text := fmt.Sprintf("%s%s", padArg(arg.Name), arg.Usage)
		args = append(args, strings.Trim(text, " "))
	}

	var argsSection bytes.Buffer
	if err := tpl.Execute(&argsSection, args); err != nil {
		panic(err)
	}

	return argsSection.String()
}

func getArgPadder(t *runner.Task) func(string) string {
	maxLength := 0
	for _, arg := range t.Args {
		if len(arg.Name) > maxLength {
			maxLength = len(arg.Name)
		}
	}
	s := fmt.Sprintf("%%-%ds", maxLength+2)
	return func(text string) string {
		return fmt.Sprintf(s, text)
	}
}
