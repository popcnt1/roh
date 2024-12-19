#!/bin/bash

# Check for required arguments
if [ $# -lt 2 ]; then
  echo "Usage: $0 <act_file> <exp_file>"
  exit 1
fi

act_file="$1"
exp_file="$2"

# Check if files exist
if [ ! -f "$act_file" ] || [ ! -f "$exp_file" ]; then
  echo "Error: One or both files not found."
  exit 1
fi

declare -A act_dict
declare -A exp_dict

# Read actual states into an associative array
while IFS=':' read -r tx_order state || [ -n "$tx_order" ]; do
  act_dict["$tx_order"]="$state"
done < "$act_file"

# Read expected states into an associative array
while IFS=':' read -r tx_order state || [ -n "$tx_order" ]; do
  exp_dict["$tx_order"]="$state"
done < "$exp_file"

not_found_count=0
matched_count=0
mismatched_states=""

for tx_order in "${!act_dict[@]}"; do
  if [[ -z "${exp_dict[$tx_order]}" ]]; then
    not_found_count=$((not_found_count + 1))
  else
    if [[ "${act_dict[$tx_order]}" != "${exp_dict[$tx_order]}" ]]; then
      mismatched_states+="$tx_order, act: ${act_dict[$tx_order]}, exp: ${exp_dict[$tx_order]}\n"
    else
      matched_count=$((matched_count + 1))
    fi
  fi
done

if [[ -n "$mismatched_states" ]]; then
  echo "Mismatched states:"
  echo -e "$mismatched_states"
else
  echo "Matched: $matched_count ($not_found_count not found)"
fi
