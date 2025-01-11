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
declare -A act_acc_dict
declare -A exp_dict
declare -A exp_acc_dict

# Read actual states (state_root + tx_accumulator_root) into associative arrays
while IFS=':' read -r tx_order state_root tx_accumulator_root || [ -n "$tx_order" ]; do
  act_dict["$tx_order"]="$state_root"
  act_acc_dict["$tx_order"]="$tx_accumulator_root"
done < "$act_file"

# Read expected states (state_root + tx_accumulator_root) into associative arrays
while IFS=':' read -r tx_order state_root tx_accumulator_root || [ -n "$tx_order" ]; do
  exp_dict["$tx_order"]="$state_root"
  exp_acc_dict["$tx_order"]="$tx_accumulator_root"
done < "$exp_file"

not_found_count=0
matched_count=0
mismatched_states=""

for tx_order in "${!act_dict[@]}"; do
  if [[ -z "${exp_dict[$tx_order]}" ]]; then
    # tx_order not found in the expected file
    not_found_count=$((not_found_count + 1))
  else
    # Compare state_root and tx_accumulator_root
    if [[ "${act_dict[$tx_order]}" != "${exp_dict[$tx_order]}" || \
          "${act_acc_dict[$tx_order]}" != "${exp_acc_dict[$tx_order]}" ]]; then

      mismatched_states+="$tx_order, act: {state_root: ${act_dict[$tx_order]}, tx_accumulator_root: ${act_acc_dict[$tx_order]}}, "
      mismatched_states+="exp: {state_root: ${exp_dict[$tx_order]}, tx_accumulator_root: ${exp_acc_dict[$tx_order]}}\n"
    else
      # Everything matches
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