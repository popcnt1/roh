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
  # Send request and extract state_root
  hash=$(rooch rpc request --method rooch_getTransactionsByOrder --params "[$current]" | jq -r '.data[0].execution_info.state_root')
  
  # Check if the return value is empty
  if [ -z "$hash" ]; then
    echo "$current:<no_hash>"
  else
    echo "$current:$hash"
  fi

  # Increment by interval
  current=$((current + interval))
done

# Ensure the end value is requested once (if not a multiple of interval)
if [ $(( (end - start) % interval )) -ne 0 ]; then
  hash=$(rooch rpc request --method rooch_getTransactionsByOrder --params "[$end]" | jq -r '.data[0].execution_info.state_root')
  if [ -z "$hash" ]; then
    echo "$end:<no_hash>"
  else
    echo "$end:$hash"
  fi
fi
