#!/bin/bash

_tusk_bash_autocomplete() {
    local cur opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="$( "${COMP_WORDS[@]:0:$COMP_CWORD}" --generate-bash-completion )"
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    return 0
}

complete -F _tusk_bash_autocomplete tusk
