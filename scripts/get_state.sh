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

# Fetch transactions and process data
result=$(rooch transaction get-transactions-by-order --cursor "$start" --limit "$limit" -d false)

state_roots=$(echo "$result" | jq -r '.data[].execution_info.state_root')
tx_accumulator_roots=$(echo "$result" | jq -r '.data[].transaction.sequence_info.tx_accumulator_root')

# Split results into arrays robustly
mapfile -t state_root_array < <(echo "$state_roots")
mapfile -t tx_accumulator_root_array < <(echo "$tx_accumulator_roots")

# Ensure the counts match
if [ ${#state_root_array[@]} -ne ${#tx_accumulator_root_array[@]} ]; then
  echo "Error: Mismatched data arrays received. Aborting process."
  exit 1
fi

count=0
for state_root in "${state_root_array[@]}"; do
  actual_cursor=$((start + count))

  tx_accumulator_root=${tx_accumulator_root_array[count]}

  # Stop process if either value is missing
  if [ -z "$state_root" ] || [ -z "$tx_accumulator_root" ]; then
    echo "Error: Could not retrieve required roots at tx_order: $actual_cursor"
    exit 1
  fi

  # Output the results in the desired format
  echo "$actual_cursor:$state_root:$tx_accumulator_root"

  count=$((count + 1))
done