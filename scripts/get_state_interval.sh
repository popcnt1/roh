#!/bin/bash

# Check the number of arguments
if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <start> <end> <interval>"
  exit 1
fi

# Get script arguments
start=$1
end=$2
interval=$3

# Check if the arguments are numbers
if ! [[ "$start" =~ ^[0-9]+$ ]] || ! [[ "$end" =~ ^[0-9]+$ ]] || ! [[ "$interval" =~ ^[0-9]+$ ]]; then
  echo "Error: All parameters must be positive integers."
  exit 1
fi

# Ensure that start is less than or equal to end
if [ "$start" -gt "$end" ]; then
  echo "Error: start must be less than or equal to end."
  exit 1
fi

# Loop to send requests and process results
current=$start
while [ "$current" -le "$end" ]; do
  # Send request and extract state_root and tx_accumulator_root
  result=$(rooch transaction get-transactions-by-order --cursor "$current" --limit 1 -d false)
  state_root=$(echo "$result" | jq -r '.data[0].execution_info.state_root')
  tx_accumulator_root=$(echo "$result" | jq -r '.data[0].transaction.sequence_info.tx_accumulator_root')

  # Stop process if either state_root or tx_accumulator_root is empty
  if [ -z "$state_root" ] || [ -z "$tx_accumulator_root" ]; then
    echo "Error: Could not retrieve required roots at tx_order: $current"
    exit 1
  fi

  # Output the results in the desired format
  echo "$current:$state_root:$tx_accumulator_root"

  # Increment by interval
  current=$((current + interval))
done

# Ensure the end value is requested once (if not a multiple of interval)
if [ $(( (end - start) % interval )) -ne 0 ]; then
  result=$(rooch transaction get-transactions-by-order --cursor "$end" --limit 1 -d false)
  state_root=$(echo "$result" | jq -r '.data[0].execution_info.state_root')
  tx_accumulator_root=$(echo "$result" | jq -r '.data[0].transaction.sequence_info.tx_accumulator_root')

  if [ -z "$state_root" ] || [ -z "$tx_accumulator_root" ]; then
    echo "Error: Could not retrieve required roots at tx_order: $end"
    exit 1
  fi

  echo "$end:$state_root:$tx_accumulator_root"
fi