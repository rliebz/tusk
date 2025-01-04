package appcli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/urfave/cli"

	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
)

// init sets the help templates for urfave/cli.
func init() { //nolint: gochecknoinits
	// These are both used, so both must be overridden
	cli.HelpPrinterCustom = wrapPrinter(cli.HelpPrinterCustom)
	cli.HelpPrinter = func(w io.Writer, templ string, data any) {
		cli.HelpPrinterCustom(w, templ, data, nil)
	}

	cli.FlagNamePrefixer = flagPrefixer

	cli.AppHelpTemplate = `{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
   {{ if .UsageText }}
   {{- .UsageText }}
   {{- else }}
   {{- .HelpName }}
   {{- if .VisibleFlags }} [global options]{{ end }}
   {{- if .Commands }} <task> [task options]{{ end }}
   {{- if .ArgsUsage }} {{ .ArgsUsage }}{{ end }}
   {{- end }}

{{- if and .Version (not .HideVersion) }}

Version:
   {{ .Version }}

{{- end }}

{{- if .Description }}

Description:
{{ indent 3 .Description }}
{{- end }}

{{- if .VisibleCommands }}

Tasks:
{{- range .VisibleCategories }}
{{- $categoryName := .Name }}
{{- with $categoryName }}
   {{ . }}:
{{- end }}
{{- range .VisibleCommands }}
   {{ if $categoryName }}  {{ end }}{{ join .Names ", " }}{{ "\t" }}{{ .Usage }}
{{- end }}
{{- end }}

{{- end }}

{{- if .VisibleFlags }}

Global Options:
{{- range .VisibleFlags }}
   {{ . }}
{{- end }}

{{- end }}

{{- with .Copyright }}

Copyright:
   {{ . }}

{{- end }}
`
}

// ShowAppHelp shows the help for a given app.
func ShowAppHelp(logger *ui.Logger, app *cli.App) {
	app.Setup()
	cli.HelpPrinter(logger.Stdout, cli.AppHelpTemplate, app)
}

type helpPrinterCustom = func(io.Writer, string, any, map[string]any)

// helpPrinter includes the custom indent template function.
func wrapPrinter(p helpPrinterCustom) helpPrinterCustom {
	return func(w io.Writer, tpl string, data any, funcs map[string]any) {
		customFuncs := map[string]any{
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

func createCommandHelp(
	command *cli.Command,
	t *runner.Task,
	dependencies []*runner.Option,
) string {
	return fmt.Sprintf(`{{.HelpName}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
{{- if .UsageText }}
   {{ .UsageText }}
{{- else }}
   {{ .HelpName }}
   {{- if .VisibleFlags }} [options]{{ end }}
   {{- with .ArgsUsage }} {{ . }}{{ end }}
{{- end }}

{{- with .Category }}

Category:
   {{ . }}

{{- end }}

{{- with .Description }}

Description:
{{ indent 3 . }}

{{- end }}%s
`, createArgsSection(t)+createOptionsSection(command, t, dependencies))
}

func createArgsSection(t *runner.Task) string {
	argsTpl := `{{- if . }}

Arguments:
{{- range  . }}
   {{ . }}
{{- end }}

{{- end }}`

	tpl := template.New(fmt.Sprintf("%s help", t.Name))
	tpl = template.Must(tpl.Parse(argsTpl))

	width := maxArgWidth(t) + 2

	lines := make([]string, 0, len(t.Args))
	for _, arg := range t.Args {
		text := fmt.Sprintf("%s%s", pad(arg.Name, width), arg.Usage)
		lines = append(lines, strings.TrimSpace(text))
	}

	var argsSection bytes.Buffer
	if err := tpl.Execute(&argsSection, lines); err != nil {
		panic(err)
	}

	return argsSection.String()
}

func maxArgWidth(t *runner.Task) int {
	maxWidth := 0
	for _, arg := range t.Args {
		maxWidth = max(maxWidth, len(arg.Name))
	}
	return maxWidth
}

func createOptionsSection(
	command *cli.Command,
	t *runner.Task,
	dependencies []*runner.Option,
) string {
	tpl := template.New(fmt.Sprintf("%s help", command.Name))
	tpl = template.Must(tpl.Parse(`{{- if . }}

Options:
{{- range  . }}
   {{ . }}
{{- end }}

{{- end }}`))

	width := maxOptionWidth(command, dependencies) + 2

	lines := make([]string, 0, len(t.Args))
	for _, flag := range command.VisibleFlags() {
		opt := optionForFlag(flag, dependencies)
		line := pad(opt.FlagText(), width) + opt.Usage
		if defaultValue, ok := opt.StaticDefault(); ok {
			if opt.Usage != "" {
				line += " "
			}
			line += fmt.Sprintf("(default: %s)", defaultValue)
		}
		lines = append(lines, strings.TrimRight(line, " "))
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, lines); err != nil {
		panic(err)
	}

	return buf.String()
}

func maxOptionWidth(command *cli.Command, opts []*runner.Option) int {
	maxWidth := 0
	for _, flag := range command.VisibleFlags() {
		opt := optionForFlag(flag, opts)
		maxWidth = max(maxWidth, len(opt.FlagText()))
	}

	return maxWidth
}

func optionForFlag(f cli.Flag, opts []*runner.Option) *runner.Option {
	flagName, _, _ := strings.Cut(f.GetName(), ",")
	for _, opt := range opts {
		if opt.Name == flagName {
			return opt
		}
	}

	panic("failed to find opt for flag: " + flagName)
}

// pad a string to a given width.
func pad(text string, width int) string {
	s := fmt.Sprintf("%%-%ds", width)
	return fmt.Sprintf(s, text)
}
