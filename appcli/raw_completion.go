package appcli

const (
	rawZshCompletion = `#compdef tusk

local meta end
local -a _words _options

let end=$CURRENT-1
IFS=$'\n' _words=( $(${words[@]:0:$end} --generate-bash-completion) )

# Split words into completion type and options
meta="${_words[1]}"
_options=( "${_words[@]:1}" )

__tusk_tasks() {
    local -a tasks
    for option in "${_options[@]}"; do
        if [[ ! "${option}" = --* ]]; then
            tasks+=("${option}")
        fi
    done
    _describe -t tasks 'tasks' tasks
}

__tusk_task_args() {
    local -a args
    for option in "${_options[@]}"; do
        if [[ "${option}" != --* ]]; then
            args+=("${option}")
        fi
    done
    if [[ ${#args} == 0 ]]; then
        _files
    else
        _describe -t task-args 'task arguments' args
    fi
}

__tusk_task_flags() {
    local -a flags
    for option in "${_options[@]}"; do
        if [[ "${option}" = --* ]]; then
            flags+=("${option}")
        fi
    done
    _describe -t task-flags 'task options' flags
}

__tusk_global_flags() {
    local -a flags
    for option in "${_options[@]}"; do
        if [[ "${option}" = --* ]]; then
            flags+=("${option}")
        fi
    done
    _describe -t global-flags 'global options' flags
}

__tusk_values() {
    local -a values
    for option in "${_options[@]}"; do
        if [[ ! "${option}" = --* ]]; then
            values+=("${option}")
        fi
    done
    _describe -t values 'values' values
}

case "${meta}" in
    "normal")
        _alternative \
            'tasks:task:__tusk_tasks' \
            'global-flags:flag:__tusk_global_flags'
        ;;
    "task-args")
        _alternative \
            'task-args:arg:__tusk_task_args' \
            'task-flags:flag:__tusk_task_flags'
        ;;
    "task-no-args")
        _alternative \
            'task-flags:flag:__tusk_task_flags'
        ;;
    "value")
        _alternative \
            'values:value:__tusk_values'
        ;;
    "file")
        _files
        ;;
esac
`

	rawBashCompletion = `#!/bin/bash

_tusk_bash_autocomplete() {
    local cur words opts meta
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    words="$( "${COMP_WORDS[@]:0:$COMP_CWORD}" --generate-bash-completion \
        | sed 's/\\:/_=_/g' | cut -f1 -d":" | sed 's/_=_/:/g' )"

    # Split words into completion type and options
    meta="$( echo "${words}" | head -n1 )"
    opts="$( echo "${words}" | tail -n +2 )"

    case "${meta}" in
        file)
            COMPREPLY=( $(compgen -f -- "${cur}") )
            ;;
        *)
            declare -a values args flags
            values=( ${opts} )
            for option in "${values[@]}"; do
                if [[ "${option}" = --* ]]; then
                    flags+=("${option}")
                else
                    args+=("${option}")
                fi
            done

            if [[ "${cur}" = --* ]]; then
                COMPREPLY=( $(compgen -W "${flags[*]}" -- "${cur}") )
            else
                COMPREPLY=( $(compgen -W "${args[*]}" -- "${cur}") )
            fi
            ;;
    esac

    return 0
}

complete -o filenames -o bashdefault -F _tusk_bash_autocomplete tusk
`
)
