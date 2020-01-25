## The Spec

### Tasks

The core of every `tusk.yml` file is a list of tasks. Tasks are declared at the
top level of the `tusk.yml` file and include a list of tasks.

For the following tasks:

```yaml
tasks:
  hello:
    run: echo "Hello, world!"
  goodbye:
    run: echo "Goodbye, world!"
```

The commands can be run with no additional configuration:

```text
$ tusk hello
Running: echo "Hello, world!"
Hello, world!
```

Tasks can be documented with a one-line `usage` string and a slightly longer
`description`. This information will be displayed in help messages:

```yaml
tasks:
  hello:
    usage: Say hello to the world
    description: |
      This command will echo "Hello, world!" to the user. There's no
      surprises here.
    run: echo "Hello, world!"
  goodbye:
    run: echo "Goodbye, world!"
```

### Run

The behavior of a task is defined in its `run` clause. A `run` clause can be
used for commands, sub-tasks, or setting environment variables. Although each
`run` item can only perform one of these actions, they can be run in succession
to handle complex scenarios.

In its simplest form, `run` can be given a string or list of strings to be
executed serially as shell commands:

```yaml
tasks:
  hello:
    run: echo "Hello!"
```

This is a shorthand syntax for the following:

```yaml
tasks:
  hello:
    run:
      - command:
          exec: echo "Hello!"
```

The `run` clause tasks a list of `run` items, which allow executing shell
commands with `command`, setting or unsetting environment variables with
`set-environment`, running other tasks with `task`, and controlling conditional
execution with `when`.

#### Command

The `command` clause is the most common thing to do during a `run`, so for
convenience, passing a string or single item will be correctly interpreted.
Here are several examples of equivalent `run` clauses:

```yaml
run: echo "Hello!"

run:
  - echo "Hello!"

run:
  command: echo "Hello!"

run:
  - command: echo "Hello!"

run:
  - command:
      exec: echo "Hello!"
```

##### Exec

The `exec` clause contains the actual shell command to be performed.

If any of the run commands execute with a non-zero exit code, Tusk will
immediately exit with the same exit code without executing any other commands.

Commands are executed using the `$SHELL` environment variable, defaulting to
`sh`. Each command in a `run` clause gets its own sub-shell, so things like
declaring functions and environment variables will not be available across
separate run commmands, although it is possible to run the `set-environment`
clause or use a multi-line shell command.

For multi-line shell commands, to preserve the exit-on-error behavior, it is
recommend to run `set -e` at the top of the script, much like any shell script.

```yaml
tasks:
  hello:
    run: |
      set -e
      errcho() {
        >&2 echo "$@"
      }
      errcho "Hello, world!"
      errcho "Goodbye, world!"
```

##### Print

Sometimes it may not be desirable to print the exact command run, for example,
if it's overly verbose or contains secrets. In that case, the `command` clause
can be passed a `print` string to use as an alternative:

```yaml
tasks:
  hello:
    run:
      command:
        exec: echo "SECRET_VALUE"
        print: echo "*****"
```

##### Dir

The `dir` clause sets the working directory for a specific command:

```yaml
tasks:
  hello:
    run:
      command:
        exec: echo "Hello from $PWD!"
        dir: ./subdir
```

#### Set Environment

To set or unset environment variables, simply define a map of environment
variable names to their desired values:

```yaml
tasks:
  hello:
    options:
      proxy-url:
        default: http://proxy.example.com
    run:
      - set-environment:
          http_proxy: ${proxy-url}
          https_proxy: ${proxy-url}
          no_proxy: ~
      - command: curl http://example.com
```

Passing `~` or `null` to an environment variable will explicitly unset it,
while passing an empty string will set it to an empty string.

Environment variables once modified will persist until Tusk exits.

#### Sub-Tasks

Run can also execute previously-defined tasks:

```yaml
tasks:
  one:
    run: echo "Inside one"
  two:
    run:
      - task: one
      - command: echo "Inside two"
```

For any arg or option that a sub-task defines, the parent task can pass a
value, which is treated the same way as passing by command-line would be. Args
are passed in as a list, while options are a map from flag name to value.

To pass values, use the long definition of a sub-task:

```yaml
tasks:
  greet:
    args:
      name:
        usage: The person to greet
    options:
      greeting:
        default: Hello
    run: echo "${greeting}, ${person}!"
  greet-myself:
    run:
      task:
        name: greet
        args:
          - me
        options:
          greeting: Howdy
```

In cases where a sub-task may not be useful on its own, define it as private to
prevent it from being invoked directly from the command-line. For example:

```yaml
tasks:
  configure-environment:
    private: true
    run:
      set-environment: {APP_ENV: dev}
  serve:
    run:
      - task: configure-environment
      - command: python main.py
```

### When

For conditional execution, `when` clauses are available.

```yaml
run:
  when:
    os: linux
  command: echo "This is a linux machine"
```

In a `run` clause, any item with a true `when` clause will execute. There are
five different checks supported:

- `command` (list): Execute if any command runs with an exit code of `0`.
  Commands will execute in the order defined and stop execution at the first
  successful command.
- `exists` (list): Execute if any of the listed files exists.
- `not-exists` (list): Execute if any of the listed files doesn't exist.
- `os` (list): Execute if the operating system matches any one from the list.
- `environment` (map[string -> list]): Execute if the environment variable
  matches any of the values it maps to. To check if a variable is not set, the
  value should be `~` or `null`.
- `equal` (map[string -> list]): Execute if the given option equals any of the
  values it maps to.
- `not-equal` (map[string -> list]): Execute if the given option is not equal to
  any one of the values it maps to.

The `when` clause supports any number of different checks as a list, where each
check must pass individually for the clause to evaluate to true. Here is a more
complicated example of how `when` can be used:

```yaml
tasks:
  echo:
    options:
      cat:
        usage: Cat a file
    run:
      - when:
          os:
            - linux
            - darwin
        command: echo "This is a unix machine"
      - when:
          - exists: my_file.txt
          - equal: {cat: true}
          - command: command -v cat
        command: cat my_file.txt
```

#### Short Form

Because it's common to check if a boolean flag is set to true, `when` clauses
also accept strings as shorthand. Consider the following example, which checks
to see if some option `foo` has been set to `true`:

```yaml
when:
  equal: {foo: true}
```

This can be expressed more succinctly as the following:

```yaml
when: foo
```

#### When Any/All Logic

A `when` clause takes a list of items, where each item can have multiple checks.
Each `when` item will pass if _any_ of the checks pass, while the whole clause
will only pass if _all_ of the items pass. For example:

```yaml
tasks:
  exists:
    run:
      - when:
          # There is a single `when` item with two checks
          exists:
            - file_one.txt
            - file_two.txt
        command: echo "At least one file exists"
      - when:
          # There are two separate `when` items with one check each
          - exists: file_one.txt
          - exists: file_two.txt
        command: echo "Both files exist"
```

These properties can be combined for more complicated logic:

```yaml
tasks:
  echo:
    options:
      verbose:
        type: bool
      ignore-os:
        type: bool
    run:
      - when:
          # (OS is linux OR darwin OR ignore OS is true) AND (verbose is true)
          - os:
              - linux
              - darwin
            equal: {ignore-os: true}
          - equal: {verbose: true}
        command: echo "This is a unix machine"
```

### Args

Tasks may have args that are passed directly as inputs. Any arg that is defined
is required for the task to execute.

```yaml
tasks:
  greet:
    args:
      name:
        usage: The person to greet
    run: echo "Hello, ${name}!"
```

The task can be invoked as such:

```text
$ tusk greet friend
Hello, friend!
```

#### Arg Values

Args can specify which values are considered valid:

```yaml
tasks:
  greet:
    args:
      name:
        values:
          - Abby
          - Bobby
          - Carl
```

Any value passed by command-line must be one of the listed values, or the
command will fail to execute.

### Options

Tasks may have options that are passed as GNU-style flags. The following
configuration will provide `-n, --name` flags to the CLI and help documentation,
which will then be interpolated:

```yaml
tasks:
  greet:
    options:
      name:
        usage: The person to greet
        short: n
        environment: GREET_NAME
        default: World
    run: echo "Hello, ${name}!"
```

The above configuration will evaluate the value of `name` in order of highest
priority:

1. The value passed by command line flags (`-n` or `--name`)
2. The value of the environment variable (`GREET_NAME`), if set
3. The value set in default

For short flag names, values can be combined such that `tusk foo -ab` is exactly
equivalent to `tusk foo -a -b`.

#### Option Types

Options can be of the types `string`, `integer`, `float`, or `boolean`, using
the zero-value of that type as the default if not set. Options without types
specified are considered strings.

For boolean values, the flag should be passed by command line without any
arugments. In the following example:

```yaml
tasks:
  greet:
    options:
      loud:
        type: bool
    run:
      - when:
          equal: {loud: true}
        command: echo "HELLO!"
      - when:
          equal: {loud: false}
        command: echo "Hello."
```

The flag should be passed as such:

```bash
tusk greet --loud
```

This means that for an option that is true by default, the only way to disable
it is with the following syntax:

```bash
tusk greet --loud=false
```

Of course, options can always be defined in the reverse manner to avoid this
issue:

```yaml
options:
  no-loud:
    type: bool
```

#### Option Defaults

Much like `run` clauses accept a shorthand form, passing a string to `default`
is shorthand. The following options are exactly equivalent:

```yaml
options:
  short:
    default: foo
  long:
    default:
      - value: foo
```

A `default` clause can also register the `stdout` of a command as its value:

```yaml
options:
  os:
    default:
      command: uname -s
```

A `default` clause also accepts a list of possible values with a corresponding
`when` clause. The first `when` that evaluates to true will be used as the
default value, with an omitted `when` always considered true.

In this example, linux users will have the name `Linux User`, while the default
for all other OSes is `User`:

```yaml
options:
  name:
    default:
      - when:
          os: linux
        value: Linux User
      - value: User
```

#### Option Values

Like args, an option can specify which values are considered valid:

```yaml
options:
  number:
    default: zero
    values:
      - one
      - two
      - three
```

Any value passed by command-line flags or environment variables must be one of
the listed values. Default values, including commands, are excluded from this
requirement.

#### Required Options

Options may be required if there is no sane default value. For a required flag,
the task will not execute unless the flag is passed:

```yaml
options:
  file:
    required: true
```

A required option cannot be private or have any default values.

#### Private Options

Sometimes it may be desirable to have a variable that cannot be directly
modified through command-line flags. In this case, use the `private` option:

```yaml
options:
  user:
    private: true
    default:
      command: whoami
```

A private option will not accept environment variables or command line flags,
and it will not appear in the help documentation.

#### Shared Options

Options may also be defined at the root of the config file to be shared between
tasks:

```yaml
options:
  name:
    usage: The person to greet
    default: World

tasks:
  hello:
    run: echo "Hello, ${name}!"
  goodbye:
    run: echo "Goodbye, ${name}!"
```

Any shared variables referenced by a task will be exposed by command-line when
invoking that task. Shared variables referenced by a sub-task will be evaluated
as needed, but not exposed by command-line.

Tasks that define an argument or option with the same name as a shared task will
overwrite the value of the shared option for the length of that task, not
including sub-tasks.

### Finally

The `finally` clause is run after a task's `run` logic has completed, whether or
not that task was successful. This can be useful for clean-up logic. A `finally`
clause has the same format as a `run` clause:

```yaml
tasks:
  hello:
    run:
      - echo "Hello"
      - exit 1          # `run` clause stops here
      - echo "Oops!"    # Never prints
    finally:
      - echo "Goodbye"  # Always prints
      - task: cleanup
  # ...
```

If the `finally` clause runs an unsuccessful command, it will terminate early
the same way that a `run` clause would. The exit code is still passed back to
the command line. However, if both the `run` clause and `finally` clause fail,
the exit code from the `run` clause takes precedence.

### Include

In some cases it may be desirable to split the task definition into a separate
file. The `include` clause serves this purpose. At the top-level of a task, a
task may optionally be specified using just the `include` key, which maps to a
separate file where there task definition is stored.

For example, `tusk.yml` could be written like this:

```yaml
tasks:
  hello:
    include: .tusk/hello.yml
```

With a `.tusk/hello.yml` that looks like this:

```yaml
options:
  name:
    usage: The person to greet
    default: World
run: echo "Hello, ${name}!"
```

It is invalid to split the configuration; if the `include` clause is used, no
other keys can be specified in the `tusk.yml`, and the full task must be
defined in the included file.

### CLI Metadata

It is also possible to create a custom CLI tool for use outside of a project's
directory by using shell aliases:

```bash
alias mycli="tusk -f /path/to/tusk.yml"
```

In that case, it may be useful to override the tool name and usage text that
are provided as part of the help documentation:

```yaml
name: mycli
usage: A custom aliased command-line application

tasks:
  ...
```

The example above will produce the following help documentation:

```text
mycli - A custom aliased command-line application

Usage:
  mycli [global options] <task> [task options]

Tasks:
  ...
```

### Interpolation

The interpolation syntax for a variable `foo` is `${foo}`, meaning any instances
of `${foo}` in the configuration file will be replaced with the value of `foo`
during execution.

Interpolation is done on a task-by-task basis, meaning args and options defined
in one task will not interpolate to any other tasks. Shared options, on the
other hand, will only be evaluated once per execution.

The execution order is as followed:

1. Shared options are interpolated first, in the order defined by the config
   file. The results of global interpolation are cached and not re-run.
2. The args for the current task being run are interpolated, in order.
3. The options for the current task being run are interpolated, in order.
4. For each call to a sub-task, the process is repeated, ignoring the task-
   specific interpolations for parent tasks, using the cached shared options.

This means that options can reference other options or args:

```yaml
options:
  name:
    default: World
  greeting:
    default: Hello, ${name}

tasks:
  greet:
    run: echo "${greeting}"
```

Because interpolation is not always desirable, as in the case of environment
variables, `$$` will escape to `$` and ignore interpolation. It is also
possible to use alternative syntax such as `$foo` to avoid interpolation as
well. The following two tasks will both use environment variables and not
attempt interpolation:

```yaml
tasks:
  one:
    run: Hello, $${USER}
  two:
    run: Hello, $USER
```

Interpolation works by substituting the value in the `yaml` config file, then
parsing the file after interpolation. This means that variable values with
newlines or other characters that are relevant to the `yaml` spec or the `sh`
interpreter will need to be considered by the user. This can be as simple as
using quotes when appropriate.
