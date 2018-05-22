# Changelog
This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
### Removed
- Remove deprecated `not_equal` syntax in favor of `not-equal`.


## 0.3.5 (2018-04-11)
### Added
- Positional arguments for tasks are now supported. All positional arguments
  specified in the config file are required.

### Changed
- Tagline is now "the modern task runner".
- Help documentation for flags with placeholder options now display them using
  angular brackets.
- Minor changes to certain error messages.

## 0.3.4 (2018-03-14)
### Added
- Top-level keys `name` and `usage` can be changed in the config file to update
  help documentation. This supports aliasing different config files to be used
  as custom CLI tools.
- Support `tusk.yaml` as alternate config file name.

## 0.3.3 (2018-02-21)
### Added
- Support environment variable conditional in `when` clauses.

### Changed
- Warnings now appear more consistently for deprecated functionality.
- Help documentation for flags is now more consistently structured.
- Several configuration keys have been renamed. While the original names are
  still supported, they have been deprecated and will be removed in a future
  release. See the deprecated section for details.

### Deprecated
- The key name `not_equal` should be replaced with `not-equal`. This change is
  to reinforce the convention for naming multi-word keys such as flag names
  using kebab case.
- The key name `environment` in `run` clauses should be replaced with
  `set-environment`. This is to make the behavior distinction from other
  `environment` clauses clear.


## 0.3.2 (2018-01-23)
### Added
- Support lists of `when` items.

### Deprecated
- Individual `when` items should no longer contain multiple validations for AND
  logic. Multiple validations for OR logic will be added in a future release.


## 0.3.1 (2018-01-05)
### Fixed
- Invoking the same sub-task multiple times with different options now assigns
  the options to each sub-task correctly.
- Short flags with arguments now offer Bash/Zsh completions correctly.


## 0.3.0 (2018-01-04)
### Changed
- Interpolation is now done per task. This has the following effects:
  - Sub-task options are no longer exposed to the command line.
  - Sub-task options are now exposed in run clauses and can be passed by a
    parent task to a sub-task.
  - Shared options are only exposed when used directly by the invoked task and
    not when invoked by sub-tasks.
  - Tasks and sub-tasks can define options with the same name.
- Sub-tasks can now be defined in any order.

### Fixed
- Sub-tasks no longer execute multiple times per reference in situations where
  the same sub-task is referenced in multiple places.

### Removed
- Environment variables can no longer be set inside of an `option` clause.
  Using `environment` inside a `run` clause is the replacement behavior.


## 0.2.3 (2017-12-19)
### Fixed
- Fix issue with Zsh tab completions.

## 0.2.2 (2017-12-14)
### Added
- Add `values` field for options. Any option directly passed by command line
  flag or environment variable must be one of the listed values, if specified.
- Add completion for option values.

### Fixed
- Fix issue where global flags with hyphens are sometimes skipped during 
  interpolation.
- Fix various issues with Zsh tab completions.

## 0.2.1 (2017-11-12)
### Added
- Environment variables can now be set and unset inside of a `run` clause. This
  replaces the `export` functionality that was previously under `option`.
- Tasks can now be defined as private.

### Changed
- Change to more minimalistic UI output theme.
- Log-level messages (Debug, Info, Warning, and Error) are now printed in title
  case instead of all caps.

### Deprecated
- Environment variables should no longer be set inside of an `option` clause.
  Using `environment` inside a `run` clause is the replacement behavior.


## 0.2.0 (2017-11-08)
### Added
- Windows is now supported.
- New -s/--silent global option available for no stderr/stdout.

### Changed
- Shell commands are now executed by the `SHELL` environment variable by
  default. If `SHELL` is not set, `sh` is used.
- Commands and flags are now listed in alphabetical order.

### Fixed
- Avoid infinite loop when searching for tusk.yml in non-unix file systems.
- Remove redundant error message for non-exit exec errors.
- Improve error messaging for non-generic yaml parsing errors.
- Fix indentation for the task list in `--help` output.


## 0.1.5 (2017-10-26)
### Changed
- Include expected value for skipped run clauses during verbose logging.

### Fixed
- Bash and Zsh completions offer file completion for subcommand flags that take
  a value rather than re-offering flag names.
- The default value for numeric types is now `0`, and the default for booleans
  is now `false`. Previously, it was an empty string for both.


## 0.1.4 (2017-10-19)
### Changed
- Bash and Zsh completions now complete global flags and task flags.
- Zsh completion now includes usage information.

### Fixed
- Application no longer errors when referencing the same shared option in both
  a task and its sub-task. Redefinitions are still disallowed.


## 0.1.3 (2017-10-16)
### Added
- Completions for Bash and Zsh are now bundled with releases and automatically
  installed with Homebrew.
- Homepage is now listed for Homebrew formula.

### Changed
- Help documentation is no longer displayed during incorrect usage.

### Fixed
- Passing the help flag now prevents code execution and prints help.
- Homebrew test command is now functional.


## 0.1.2 (2017-09-26)
### Added
- Options may now be required.
- Options can be exported to environment variables.

### Changed
- Improve error handling in custom yaml unmarshalling.
- Exit for unexpected errors occuring during when clause validation.
- Use exit code 1 for all unexpected errors.
- Unexpected arguments now cause an error.
- Remove piping of command stdout/stderr to improve support for interactive
  tasks.

### Fixed
- Short names cannot exceed one character in length.


## 0.1.1 (2017-09-16)
### Changed
- The recommended way to install the latest stable version is now Homebrew or
  downloading directly from the GitHub releases page.

### Fixed
- Fix interpolation for tasks with only private options.

## 0.1.0 (2017-09-14)
### Initial Release
