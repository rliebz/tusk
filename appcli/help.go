package appcli

import "github.com/urfave/cli"

// init sets the help templates for urfave/cli.
// nolint: lll
func init() {
	cli.AppHelpTemplate = `{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} task [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

Version:
   {{.Version}}{{end}}{{end}}{{if .Description}}

Description:
   {{.Description}}{{end}}{{if len .Authors}}

Author{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

Tasks:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

Global Options:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

Copyright:
   {{.Copyright}}{{end}}
`

	cli.CommandHelpTemplate = `{{.HelpName}} - {{.Usage}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{end}}{{if .Category}}

Category:
   {{.Category}}{{end}}{{if .Description}}

Description:
   {{.Description}}{{end}}{{if .VisibleFlags}}

Options:
   {{range  $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`

	cli.SubcommandHelpTemplate = `{{.HelpName}} - {{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}
Options:
   {{range  $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`
}
