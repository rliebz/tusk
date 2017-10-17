#!/bin/bash

_tusk_bash_autocomplete() {
    local cur words opts meta
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    words="$( "${COMP_WORDS[@]:0:$COMP_CWORD}" --generate-bash-completion )"

    # Split words into completion type and options
    meta="$( echo "${words}" | head -n1 )"
    opts="$( echo "${words}" | tail -n +2 )"

    case "${meta}" in
        file)
            COMPREPLY=( $(compgen -f -- "${cur}") )
            ;;
        tasks)
            COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
            ;;
    esac

    return 0
}

complete -o filenames -o bashdefault -F _tusk_bash_autocomplete tusk
