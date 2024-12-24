package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"roh/pkg/block"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return
	}
	r := runner{
		cfg:        cfg,
		heightHash: make(map[uint64][]string),
	}
	if err := r.run(); err != nil {
		fmt.Println("Error running runner:", err)
		return
	}
}

type runner struct {
	cfg        cmdConfig
	heightHash map[uint64][]string
}

func (r *runner) run() error {

	log.Printf("Loading block idx from file %s\n", r.cfg.idxPath)
	indexer := block.NewBTCBatchIndexer()
	err := indexer.LoadFromFile(r.cfg.idxPath)
	if err != nil {
		fmt.Println("Error load block idx from file:", err)
		return err
	}
	reorg := indexer.FindReorg()
	log.Printf("Downloading blocks from %s\n", r.cfg.url)
	for height, idxes := range reorg {
		for _, idx := range idxes {
			fmt.Printf("%d,%s,%d\n", height, idx.BTCBlockHash, idx.RoochBatchID)
		}
	}
	downloaded, err := r.downloadBlocks(reorg)
	log.Printf("Downloaded %d blocks\n", downloaded)
	return err
}

func (r *runner) downloadBlocks(reorgs map[uint64][]block.BTCBatchIdx) (int, error) {

	downloaded := 0
	for height, idxes := range reorgs {
		for _, idx := range idxes {
			url := fmt.Sprintf("%s/rawblock/%s?format=hex", r.cfg.url, idx.BTCBlockHash)
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("failed to download block %s, height %d: %v", idx.BTCBlockHash, height, err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				log.Printf("failed to download block %s, height %d: status_code: %d", idx.BTCBlockHash, height, resp.StatusCode)
				continue
			}
			output := fmt.Sprintf("%s/%s", r.cfg.output, idx.BTCBlockHash)
			f, err := os.Create(output)
			if err != nil {
				return 0, fmt.Errorf("failed to create file: %v", err)
			}
			if _, err := io.Copy(f, resp.Body); err != nil {
				f.Close()
				return 0, fmt.Errorf("failed to write block to file: %v", err)
			}
			err = f.Sync()
			if err != nil {
				resp.Body.Close()
				f.Close()
				return 0, fmt.Errorf("failed to sync file: %v", err)
			}
			resp.Body.Close()
			f.Close()
			downloaded++
		}
	}
	return downloaded, nil
}

type cmdConfig struct {
	output  string
	idxPath string
	url     string
}

func parseFlags() (cfg cmdConfig, err error) {
	output := flag.String("output", "", "Output path")
	idxPath := flag.String("idx", "", "Index path")
	url := flag.String("url", "", "Blocks download URL")
	flag.Parse()
	if *output == "" || *idxPath == "" {
		fmt.Println("Usage: --output=<output> --url=<url> --idx=<index>")
		return cmdConfig{}, fmt.Errorf("missing required flags")
	}
	if *url == "" {
		*url = "https://blockchain.info"
	}

	return cmdConfig{
		output:  *output,
		idxPath: *idxPath,
		url:     *url,
	}, nil
}
