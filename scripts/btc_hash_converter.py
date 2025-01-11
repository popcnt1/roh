#!/usr/bin/env python3
import sys
import json

def bytes_to_btc_hash(bytes_array):
    hex_str = ''.join([f'{b:02x}' for b in bytes_array])
    hex_pairs = [hex_str[i:i+2] for i in range(0, len(hex_str), 2)]
    reversed_hex = ''.join(hex_pairs[::-1])
    return reversed_hex

def main():
    if len(sys.argv) != 2:
        print("Usage: ./btc_hash_converter.py '[56,50,19,157,...]'")
        sys.exit(1)
    
    try:
        bytes_array = json.loads(sys.argv[1])
        print(bytes_to_btc_hash(bytes_array))
    except json.JSONDecodeError:
        print("Error: Input must be a valid JSON array")
        sys.exit(1)

if __name__ == "__main__":
    main()
