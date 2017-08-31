# Tusk

[![CircleCI](https://img.shields.io/circleci/project/github/rliebz/tusk.svg)](https://circleci.com/gh/rliebz/tusk)
[![license](https://img.shields.io/github/license/rliebz/tusk.svg)](LICENSE)

Tusk is a yaml-based task runner. By creating a `tusk.yml` in the root of a
repository, Tusk becomes a custom command line tool with minimal configuration.

Note that as Tusk is currently unversioned, the CLI and `tusk.yml` file format
should be considered unstable and subject to change.

## Getting Started

### Installation

With a [Go][go] environment set up, simply run:

```bash
go get -u github.com/rliebz/tusk
```

### Usage

Create a `tusk.yml` file in the root of a project repository:

```yml
tasks:
  greet:
    usage: say hello to someone
    options:
      name:
        usage: a person to say "Hello" to
        default: World
    run:
      - command: echo "Hello, ${name}!"
```

As long as there is a `tusk.yml` file in the working or any parent directory,
tasks can be run:

```bash
tusk greet
```

Help messages are dynamically generated for the project and tasks:

```bash
tusk --help
tusk greet -h
```

For a more detailed example, check out [`example/example.yml`](example/example.yml)
or the project's own [`tusk.yml`](tusk.yml) file.

[go]: https://golang.org
