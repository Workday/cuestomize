#!/bin/bash

set -e

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <path-to-directory> <path-to-api>" "[dir-suffix]"
  exit 1
fi

which cue >/dev/null || { echo "cue command not found in PATH."; exit 1; }

dir_suffix=${3}

cd "$1"

for dir in */; do
  mkdir -p "$dir/${dir_suffix}"
  pushd "$dir/${dir_suffix}"

  cue mod init cue.k8s.example || true
  # clean the cue generated files
  rm -fr cue.mod/gen || true
  cue get go "$2"
  cue mod tidy || echo "cue mod tidy failed, continuing..."
  cue fmt ./... || true

  popd
done
