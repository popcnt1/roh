#!/bin/bash

# Check if a filename argument is provided
if [ $# -lt 1 ]; then
  echo "Usage: $0 <filename>"
  exit 1
fi

filename=$1

# Sort, deduplicate, and save to a temporary file (same as before)
sort -t: -k1 -n "$filename" | awk '!seen[$0]++' > "$filename.tmp"
mv "$filename.tmp" "$filename"

# Get the first and last tx_order
first_tx=$(head -n 1 "$filename" | cut -d: -f1)
last_tx=$(tail -n 1 "$filename" | cut -d: -f1)

# Get the total count of lines (after deduplication)
total_count=$(wc -l < "$filename")


echo "Range: $first_tx - $last_tx"
echo "Total count: $total_count"