function __tusk_should_complete_files
  set -l command (commandline -opc)
  set -l completions (command $command --generate-bash-completion)
  set -l meta $completions[1]

  return (string match -q -- $meta file)
end

function __tusk_list_args
  set -l command (commandline -opc)
  set -l completions (command $command --generate-bash-completion)
  set completions $completions[2..-1]

  for line in $completions
    set -l esc '{UNLIKELY_ESCAPE_SEQUENCE}'
    set -l words (
      echo $line |
      string replace --all '\:' $esc |
      string split -m1 ':' |
      string replace --all $esc ':'
    )

    if test (count $words) -gt 1
      echo $words[1]\t$words[2]
    else
      echo $words[1]
    end
  end
end

complete -F -c tusk -n '__tusk_should_complete_files'
complete -f -c tusk -n 'not __tusk_should_complete_files' -a '(__tusk_list_args)'
