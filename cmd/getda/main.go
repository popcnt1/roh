package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"roh/pkg/utils"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	output, url, resetFrom, lastChunk, maxGoroutines, err := parseFlags()
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return
	}
	chunks, err := makeChunksList(output, resetFrom, lastChunk)
	if err != nil {
		fmt.Println("Error making chunks list:", err)
		return
	}
	fmt.Printf("starting download for %d chunks\n", len(chunks))

	err = downloadChunks(chunks, url, output, maxGoroutines)
	if err != nil {
		fmt.Println("Error downloading chunks:", err)
		return
	}
}

func downloadChunks(chunks []uint64, url, output string, maxGoroutines int) error {
	baseURL := url
	tmpDir := filepath.Join(output, "tmp")

	// Ensure tmp directory exists
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	var totalSize int64
	var totalSegments uint64
	var chunkSize int64

	chunkChan := make(chan uint64, len(chunks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	doneChunk := uint64(0)

	for k := 0; k < maxGoroutines; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startTime := time.Now()
			for chunk := range chunkChan {
				doneSegments := make([]uint64, 0)
				chunkNotFound := false
				for i := uint64(0); ; i++ {
					segmentURL := fmt.Sprintf("%s/%d_%d", baseURL, chunk, i)
					resp, err := http.Get(segmentURL)
					if err != nil {
						fmt.Printf("failed to download %s: %v\n", segmentURL, err)
						doneSegments = doneSegments[:0]
						break
					}

					if resp.StatusCode == http.StatusNotFound {
						resp.Body.Close()
						// Break the loop if segment is not found
						if i == 0 {
							chunkNotFound = true
						}
						break
					}

					tmpSegmentPath := filepath.Join(tmpDir, fmt.Sprintf("%d_%d", chunk, i))
					out, err := os.Create(tmpSegmentPath)
					if err != nil {
						resp.Body.Close()
						fmt.Printf("failed to create file %s: %v\n", tmpSegmentPath, err)
						continue
					}

					// Copy the content from the response to the file
					written, err := io.Copy(out, resp.Body)
					if err != nil {
						out.Close()
						resp.Body.Close()
						fmt.Printf("failed to write data to file %s: %v\n", tmpSegmentPath, err)
						continue
					}

					out.Close()
					resp.Body.Close()
					doneSegments = append(doneSegments, i)
					mu.Lock()
					totalSize += written
					totalSegments++
					mu.Unlock()
					atomic.AddInt64(&chunkSize, written)
				}

				if chunkNotFound {
					fmt.Printf("chunk: %d not found, chunks are not continuous\n", chunk)
					break
				}
				// Move the file to the final location
				// Rename the files in descending order of segment numbers
				for i := len(doneSegments) - 1; i >= 0; i-- {
					segment := doneSegments[i]
					tmpSegmentPath := filepath.Join(tmpDir, fmt.Sprintf("%d_%d", chunk, segment))
					finalSegmentPath := filepath.Join(output, fmt.Sprintf("%d_%d", chunk, segment))
					if err := os.Rename(tmpSegmentPath, finalSegmentPath); err != nil {
						fmt.Printf("failed to move file to final location: %v\n", err)
					}
				}
				newDoneChunk := atomic.AddUint64(&doneChunk, 1)

				if newDoneChunk%500 == 0 {
					duration := time.Since(startTime)
					chunkSize500 := atomic.SwapInt64(&chunkSize, 0)
					fmt.Printf("Downloaded %d/%d chunks; last 500 chunks taken: %v; size: %s\n",
						newDoneChunk, len(chunks), duration, utils.HumanReadableBytes(chunkSize500))
					startTime = time.Now()
				}

			}
		}()
	}

	for _, chunk := range chunks {
		chunkChan <- chunk
	}
	close(chunkChan)

	wg.Wait()

	fmt.Printf("Chunks downloaded successfully\nTotal Size: %s\nTotal Segments: %d\n", utils.HumanReadableBytes(totalSize), totalSegments)
	return nil
}

func makeChunksList(output string, resetFrom, lastChunk uint64) ([]uint64, error) {
	// Search for missing chunks in output from [0, lastChunk], each chunk is made of several segments files
	// are named as output/<chunk>_<segment>
	chunks := make([]uint64, 0, lastChunk)
	for i := uint64(0); i <= lastChunk; i++ {
		path := fmt.Sprintf("%s/%d_0", output, i) // Check if segment 0 exists
		// check path exists, if not add to list
		// check if path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			chunks = append(chunks, i)
		} else if err != nil {
			// Return error if it occurs during os.Stat
			return nil, err
		}
	}
	if resetFrom > 0 {
		// append chunks from resetFrom to lastChunk
		for j := resetFrom; j <= lastChunk; j++ {
			chunks = append(chunks, j)
		}
	}
	chunks = dedup(chunks)
	// Return a list of missing chunks
	return chunks, nil
}

func dedup(chunks []uint64) []uint64 {
	seen := make(map[uint64]struct{})
	unique := make([]uint64, 0, len(chunks))
	for _, chunk := range chunks {
		if _, ok := seen[chunk]; !ok {
			seen[chunk] = struct{}{}
			unique = append(unique, chunk)
		}
	}
	return unique
}

func parseFlags() (string, string, uint64, uint64, int, error) {
	output := flag.String("output", "", "Output string")
	url := flag.String("url", "", "URL string")
	resetFrom := flag.Uint64("reset_from", 0, "Reset from chunk as uint64")
	lastChunk := flag.Uint64("last_chunk", 0, "Last chunk as uint64")
	maxThreads := flag.Int("threads", 4, "Max number of threads for downloading")
	flag.Parse()
	if *output == "" || *url == "" {
		fmt.Println("Usage: --output=<output> --url=<url> --last_chunk=<last_chunk>")
		return "", "", 0, 0, 0, fmt.Errorf("missing required flags")
	}

	return *output, *url, *resetFrom, *lastChunk, *maxThreads, nil
}
