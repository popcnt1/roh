package main

import (
	"flag"
	"fmt"
	"roh/pkg/block"
)

func main() {

	input, output, err := parseFlags()
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return
	}
	indexer := block.NewBTCBatchIndexer()
	err = indexer.MakeInMem(input)
	if err != nil {
		fmt.Println("Error making inmem:", err)
		return
	}
	err = indexer.DumpToFile(output)
	if err != nil {
		fmt.Println("Error dumping to file:", err)
		return
	}
	reorg := indexer.FindReorg()
	for _, r := range reorg {
		fmt.Printf("%v\n", r)
	}
}

func parseFlags() (string, string, error) {
	input := flag.String("input", "", "Input path")
	output := flag.String("output", "", "Output path")
	flag.Parse()
	if *output == "" || *input == "" {
		fmt.Println("Usage: --output=<output> --url=<url> --input=<input>")
		return "", "", fmt.Errorf("missing required flags")
	}

	return *input, *output, nil
}
