# Contributing

All contributions are welcome and appreciated. Feel free to open issues or pull
requests for any fixes, changes, or new features, and if you are not sure about
anything, open it anyway. Issues and pull requests are a great forum for
discussion and a great opportunity to help improve the code as a whole.

## Opening Issues

A great way to contribute to the project is to open GitHub issues whenever you
encounter any issue, or if you have an idea on how Tusk can improve. Missing
or incorrect documentation are issues too, so feel free to open one whenever
you feel there is a chance to make Tusk better.

When reporting a bug, make sure to include the expected behavior and steps to
reproduce. The more descriptive you can be, the faster the issue can be
resolved.

## Opening Pull Requests

Always feel free to open a pull request, whether it is a fix or a new addition.
For big or breaking changes, you might consider opening an issue first to check
interest, but it is absolutely not required to make a contribution.

Tests are run automatically on each PR, and 100% test and lint pass rate is
required to get the code merged in, although it is fine to have work-in-
progress pull requests open while debugging. Details on how to run the test
suite can be found [here](#running-tests).

For features which change the spec of the configuration file, documentation
should be added in [docs/spec.md][spec.md].

## Setting Up a Development Environment

For local development, you will need Go and golangci-lint installed. The
easiest way to do this is with [mise](https://mise.jdx.dev/), but it is not
required.

If it is not already on your path, you probably want to have the `GOBIN`
directory available for projects installed by `go install`. To do so, add the
following to your `.bash_profile` or `.zshrc`:

```bash
export PATH="$PATH:$(go env GOBIN)"
```

To install Tusk:

```bash
go install
```

If you have already installed tusk from another source, make sure you test
against the development version installed locally.

## Making Changes

If you have not yet done so, make sure you fork the repository so you can push
your changes back to your own fork on GitHub.

When starting work on a new feature, create a new branch based off the `main`
branch. Pull requests should generally target the `main` branch, and releases
will be cut separately.

## Running Tests

To run the unit tests:

```bash
tusk test
```

To run the full test suite, along with golangci-lint:

```bash
tusk test -a
```

If the linter fails, execution will stop short and not actually run the
unit test suite. If there is a linter error that is a false-positive, or the
violation is necessary for your contribution, you can disable a specific linter
for that line:

```golang
cmd := exec.Command("sh", "-c", command) //nolint:gosec
```

[spec.md]: https://github.com/rliebz/tusk/blob/main/docs/spec.md
