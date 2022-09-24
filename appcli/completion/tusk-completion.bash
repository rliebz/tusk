#!/bin/bash

# Unescape occurrences of "\:" in COMP_WORDS.
#
# This allows running COMP_WORDS as a command more reliably.
__unescape_comp_words_colons() {
    local i=${#COMP_WORDS[@]}
    while ((i-- > 0)); do
        COMP_WORDS[i]="${COMP_WORDS[i]//\\:/:}"
    done
}

# Remove everything up to and including the last colon in COMPREPLY.
#
# Since COMP_WORDBREAKS considers the colon to be a wordbreak by default, it
# must be removed from COMPREPLY for bash to handle it correctly.
__trim_compreply_colon_prefix() {
    local cur="$1"
    if [[ $cur != *:* || $COMP_WORDBREAKS != *:* ]]; then
        return
    fi

    local colon_prefix=${cur%"${cur##*:}"}

    local i=${#COMPREPLY[@]}
    while ((i-- > 0)); do
        COMPREPLY[i]=${COMPREPLY[i]#"$colon_prefix"}
    done
}

_tusk_bash_autocomplete() {
    local cur words opts meta
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"

    __unescape_comp_words_colons

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

    __trim_compreply_colon_prefix "$cur"

    return 0
}

complete -o filenames -o bashdefault -F _tusk_bash_autocomplete tusk
