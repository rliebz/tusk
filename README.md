# Tusk

[![GitHub release](https://img.shields.io/github/release/rliebz/tusk.svg)][releases]
[![CircleCI](https://img.shields.io/circleci/project/github/rliebz/tusk/master.svg)][circle]
[![AppVeyor](https://img.shields.io/appveyor/ci/RobertLiebowitz/tusk/master.svg?label=windows)][appveyor]
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Gitter](https://img.shields.io/gitter/room/tusk-cli/tusk.svg)][gitter]

Tusk is a yaml-based task runner. By creating a `tusk.yml` in the root of a
repository, Tusk becomes a custom command line tool with minimal configuration.

Details on the usage and configuration options can be found in the [project
documentation][documentation].

## Features

- __Customizable__: Specify your own tasks and options with support for command-line
  flags, environment variables, conditional logic, and more.
- __Explorable__: All the help you need to get started is available straight from the command line. Help documentation is generated dynamically and support for Bash and Zsh tab completion is available. 
- __Accessible__: Built for usability with a simple YAML configuration, familiar
  syntax for passing options, Bash-like variable interpolation, and a colorful
  terminal output.
- __Zero Dependencies__: All you need is a single binary file to get started on
  Linux, macOS, or Windows.

## Getting Started

### Installation

The latest version can be downloaded from the [releases page][releases].

#### Installation Script

To install automatically, or for use in CI, run the following command:

```bash
curl -sL https://git.io/tusk | bash -s -- -b /usr/local/bin latest
```

To pin to a specific version, replace `latest` with the tag for that version. This is recommended for automated scripts.

To install to another directory, change the path passed to `-b`.

#### Homebrew

On macOS, installation is also available through [homebrew][homebrew]:

```bash
brew install rliebz/tusk/tusk
```

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

```text
$ tusk greet --name friend
Running: echo "Hello, friend!"
Hello, friend!
```

Help messages are dynamically generated based on the YAML configuration:

```text
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

```text
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

[appveyor]: https://ci.appveyor.com/project/RobertLiebowitz/tusk
[circle]: https://circleci.com/gh/rliebz/tusk/tree/master
[contributing]: https://github.com/rliebz/tusk/blob/master/CONTRIBUTING.md
[documentation]: https://rliebz.github.io/tusk/
[gitter]: https://gitter.im/tusk-cli/tusk
[homebrew]: https://brew.sh
[releases]: https://github.com/rliebz/tusk/releases
[spec]: https://rliebz.github.io/tusk/spec/
[tusk.yml]: https://github.com/rliebz/tusk/blob/master/tusk.yml
