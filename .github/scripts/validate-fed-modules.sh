#!/usr/bin/env bash


# https://vaneyckt.io/posts/safer_bash_scripts_with_set_euxo_pipefail/
set -exEu -o pipefail



# Set a trap to catch errors in the script and print an error message
# with the script name and line number where the error occurred.
trap 's=$?; echo "$0: error on $0:$LINENO"; exit $s' ERR

# Find all files named "fed-modules.json" in the "static" directory
# and store their paths in an array named "files".
files=($(find static -name "fed-modules.json"))

valid=true


for file in "${files[@]}"
do

  # Read the contents of the file and pass them to the jq command to extract all keys that are not camel-cased.
  invalid_keys=$(cat $file | jq 'keys[] | select(test("^[a-z]+([A-Z][a-z]*)*$") | not)')
  
  if [ -z "$invalid_keys" ]; then
      echo "${file} is valid."
  else
      echo "${file} is invalid. Below keys must be camel-cased."
      echo "${invalid_keys}"
      valid=false
  fi

done


# If any file was found to have invalid keys, exit the script with an error code.
if ! "$valid"; then
  exit 1
fi