#!/bin/bash

# Check for required arguments
if [ $# -lt 2 ]; then
  echo "Usage: $0 <start> <limit>"
  exit 1
fi

start=$1
limit=$2

# Check if the arguments are numbers
if ! [[ "$start" =~ ^[0-9]+$ ]] || ! [[ "$limit" =~ ^[0-9]+$ ]]; then
  echo "Error: All parameters must be positive integers."
  exit 1
fi

hash_array=$(rooch transaction get-transactions-by-order --cursor "$start" --limit "$limit" -d false | jq -r '.data[].execution_info.state_root')

count=0
for hash in $hash_array; do
    actual_cursor=$((start + count))
    if [ -z "$hash" ]; then
      echo "$actual_cursor:<no_hash>"
    else
      echo "$actual_cursor:$hash"
    fi
    count=$((count + 1))
done
