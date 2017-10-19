# Changelog
This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
### Changed
- Bash and Zsh completions now complete global flags and task flags.
- Zsh completion now includes usage information.


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
