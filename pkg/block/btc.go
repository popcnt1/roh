package block

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"roh/pkg/btc"
	"sort"
	"strconv"
	"strings"
)

type BTCBatchIdx struct {
	BTCBlockHeight uint64
	BTCBlockHash   string
	RoochBatchID   uint64 // although it's uint128, but we can use uint64 for now
}

type BTCBatchIndexer struct {
	inMem map[uint64][]BTCBatchIdx
}

func NewBTCBatchIndexer() *BTCBatchIndexer {
	return &BTCBatchIndexer{
		inMem: make(map[uint64][]BTCBatchIdx),
	}
}

func (b *BTCBatchIndexer) GetBTCBatchIdx(height uint64) []BTCBatchIdx {
	return b.inMem[height]
}

func (b *BTCBatchIndexer) DumpToFile(output string) error {
	// each line is block_height,block_hash,batch_id, order by block_height
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()
	// range over inMem and write to file
	// sort by block height
	sorted := make([]uint64, 0, len(b.inMem))
	for k := range b.inMem {
		sorted = append(sorted, k)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	for _, blockHeight := range sorted {
		for _, idx := range b.inMem[blockHeight] {
			_, err = f.WriteString(fmt.Sprintf("%d,%s,%d\n", idx.BTCBlockHeight, idx.BTCBlockHash, idx.RoochBatchID))
			if err != nil {
				return fmt.Errorf("failed to write to file: %v", err)
			}
		}
	}

	return f.Sync()
}

func (b *BTCBatchIndexer) LoadFromFile(output string) error {
	f, err := os.Open(output)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	lineNum := 0
	for scanner.Scan() {
		var blockHeight uint64
		var blockHash string
		var batchID uint64
		line := scanner.Text()
		tokens := strings.Split(line, ",")
		if len(tokens) != 3 {
			log.Printf("failed to parse line: %s", line)
			continue
		}
		blockHeight, err = strconv.ParseUint(tokens[0], 10, 64)
		if err != nil {
			log.Printf("line %d: invalid block height: %s", lineNum, tokens[0])
			continue
		}

		blockHash = tokens[1] // Directly assign the block hash

		batchID, err = strconv.ParseUint(tokens[2], 10, 64)
		if err != nil {
			log.Printf("line %d: invalid Rooch Batch ID: %s", lineNum, tokens[2])
			continue
		}
		b.inMem[blockHeight] = append(b.inMem[blockHeight], BTCBatchIdx{
			BTCBlockHeight: blockHeight,
			BTCBlockHash:   blockHash,
			RoochBatchID:   batchID,
		})
		lineNum++
	}

	return nil
}

func (b *BTCBatchIndexer) FindReorg() map[uint64][]BTCBatchIdx {
	reorgs := make(map[uint64][]BTCBatchIdx)
	for height, idxs := range b.inMem {
		if len(idxs) > 1 {
			reorgs[height] = idxs
		}
	}
	return reorgs

}

func (b *BTCBatchIndexer) MakeInMem(batchDir string) error {
	height := uint64(0)
	for ; ; height++ {
		eof, err := b.parseFile(height, batchDir)
		if eof {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to parse file: %v", err)
		}
	}
	return nil
}

func (b *BTCBatchIndexer) parseFile(id uint64, batchDir string) (eof bool, err error) {
	path := fmt.Sprintf("%s/%d", batchDir, id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(scanner.Text()), &jsonData); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			continue
		}

		if data, ok := jsonData["data"].(map[string]interface{}); ok {
			if block, ok := data["L1Block"].(map[string]interface{}); ok {
				blockHeight := uint64(block["block_height"].(float64))
				hashArray := block["block_hash"].([]interface{})
				blockHash := btc.BytesToBtcHash(hashArray)
				b.inMem[blockHeight] = append(b.inMem[blockHeight], BTCBatchIdx{
					BTCBlockHeight: blockHeight,
					BTCBlockHash:   blockHash,
					RoochBatchID:   id,
				})
			}
		}
	}

	return
}
