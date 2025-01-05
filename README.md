# Tusk

[![GitHub release](https://img.shields.io/github/release/rliebz/tusk.svg)][releases]
[![Test Workflow](https://github.com/rliebz/tusk/actions/workflows/test.yml/badge.svg)](https://github.com/rliebz/tusk/actions?query=workflow%3ATest+branch%3Amain++)
[![Go Report Card](https://goreportcard.com/badge/github.com/rliebz/tusk)](https://goreportcard.com/report/github.com/rliebz/tusk)
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Tusk is a yaml-based task runner. By creating a `tusk.yml` in the root of a
repository, Tusk becomes a custom command line tool with minimal configuration.

Details on the usage and configuration options can be found in the [project
documentation][documentation].

## Features

- **Customizable**: Specify your own tasks and options with support for
  command-line flags, environment variables, conditional logic, and more.
- **Explorable**: All the help you need to get started is available straight
  from the command line. Help documentation is generated dynamically, and
  support for Bash and Zsh tab completion is available.
- **Accessible**: Built for usability with a simple YAML configuration,
  familiar syntax for passing options, Bash-like variable interpolation, and a
  colorful terminal output.
- **Zero Dependencies**: All you need is a single binary file to get started on
  Linux, macOS, or Windows.

## Getting Started

### Installation

#### Go

With Go 1.21+ installed:

```bash
go install github.com/rliebz/tusk@latest
```

#### Homebrew

On macOS, installation is also available through [homebrew][homebrew]:

```bash
brew install rliebz/tusk/tusk
```

With Homebrew, tab completion is installed automatically.

#### Compiled Releases

The latest version can be downloaded from the [releases page][releases].

To install automatically:

```bash
curl -sL https://git.io/tusk | bash -s -- -b /usr/local/bin latest
```

To pin to a specific version, replace `latest` with the tag for that version.

To install to another directory, change the path passed to `-b`.

### Installing Tab Completion

For bash:

```bash
tusk --install-completion bash
```

For fish:

```fish
tusk --install-completion fish
```

For zsh:

```zsh
tusk --install-completion zsh
```

Completions can be uninstalled with the `--uninstall-completion` flag.

### Usage

Create a `tusk.yml` file in the root of a project repository:

```yaml
tasks:
  greet:
    usage: Say hello to someone
    options:
      name:
        usage: A person to say "Hello" to
        default: World
    run: echo "Hello, ${name}!"
```

As long as there is a `tusk.yml` file in the working or any parent directory,
tasks can be run:

```console
$ tusk greet --name friend
Running: echo "Hello, friend!"
Hello, friend!
```

Help messages are dynamically generated based on the YAML configuration:

```console
$ tusk --help
tusk - the modern task runner

Usage:
   tusk [global options] <task> [task options]

Tasks:
   greet  Say hello to someone

Global Options:
   -f, --file <file>  Set file to use as the config file
   -h, --help         Show help and exit
   ...
```

Individual tasks have help messages as well:

```console
$ tusk greet --help
tusk greet - Say hello to someone

Usage:
   tusk greet [options]

Options:
   --name <value>  A person to say "Hello" to
```

Additional information on the configuration spec can be found in the [project
documentation][spec].

For more detailed examples, check out the project's own [`tusk.yml`][tusk.yml]
file.

## Contributing

Set-up instructions for a development environment and contribution guidelines
can be found in [CONTRIBUTING.md][contributing].

[contributing]: https://github.com/rliebz/tusk/blob/main/CONTRIBUTING.md
[documentation]: https://rliebz.github.io/tusk/
[homebrew]: https://brew.sh
[releases]: https://github.com/rliebz/tusk/releases
[spec]: https://rliebz.github.io/tusk/spec/
[tusk.yml]: https://github.com/rliebz/tusk/blob/main/tusk.yml
