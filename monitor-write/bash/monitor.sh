#!/bin/bash

path_to_watch="${1}"

if [ "${path_to_watch:0:1}" != "/" ]; then
  >&2 echo "Error: watched file path must be absolute"
  exit 1
fi

parent_paths_to_watch=()
current_path="${path_to_watch}"
while [ "$(dirname "${current_path}")" != "${current_path}" ]; do
  current_path="$(dirname "${current_path}")"
  parent_paths_to_watch+=("${current_path}")
done

read_contents_or_empty() {
  echo "$(cat "${path_to_watch}" 2> /dev/null | base64 --wrap=0)"
}

wait_for_event() {
  (
    shopt -s lastpipe
    inotifywait --quiet --monitor --format "%w%f" --event MOVED_FROM "${parent_paths_to_watch[@]}" | while read -r moved_path; do
      if [[ "${parent_paths_to_watch[@]}" =~ "${moved_path}" ]]; then
        exit
      fi
    done
  ) & parents_pid=$!

  (
    shopt -s lastpipe
    inotifywait --quiet --monitor --format "%w%f" --event CREATE --event CLOSE_WRITE --event MODIFY --event MOVE_SELF --event DELETE_SELF "${path_to_watch}" | while read -r changed_path; do
      if [[ "${path_to_watch}" = "${changed_path}" ]]; then
        exit
      fi
    done
  ) & child_pid=$!

  wait -n
  kill ${parents_pid} ${child_pid} 2> /dev/null
  wait
}

while true; do
  if [ -f "${path_to_watch}" ]; then
    read_contents_or_empty
    wait_for_event
  else
    read_contents_or_empty
    while [ ! -f "${path_to_watch}" ]; do
      sleep 5
    done
  fi
done
