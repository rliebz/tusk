#!/bin/bash

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
