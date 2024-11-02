package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"roh/pkg/utils"
	"strconv"
	"time"
)

func main() {
	output, url, lastChunk, err := parseFlags()
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return
	}
	chunks, err := makeChunksList(output, lastChunk)
	if err != nil {
		fmt.Println("Error making chunks list:", err)
		return
	}
	fmt.Printf("starting download for %d chunks\n", len(chunks))

	err = downloadChunks(chunks, url, output)
	if err != nil {
		fmt.Println("Error downloading chunks:", err)
		return
	}
}

func downloadChunks(chunks []uint64, url, output string) error {
	baseURL := url
	tmpDir := filepath.Join(output, "tmp")

	// Ensure tmp directory exists
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}

	var totalSize int64
	var totalSegments uint64

	startTime := time.Now()
	chunkSize := int64(0)

	for j, chunk := range chunks {
		// Download each segment from 0 to not_found_error
		segments := make([]uint64, 0)
		for i := uint64(0); ; i++ {
			segmentURL := fmt.Sprintf("%s/%d_%d", baseURL, chunk, i)
			resp, err := http.Get(segmentURL)
			if err != nil {
				return fmt.Errorf("failed to download %s: %v", segmentURL, err)
			}

			if resp.StatusCode == http.StatusNotFound {
				resp.Body.Close()
				// Break the loop if segment is not found
				break
			}

			tmpSegmentPath := filepath.Join(tmpDir, fmt.Sprintf("%d_%d", chunk, i))
			out, err := os.Create(tmpSegmentPath)
			if err != nil {
				resp.Body.Close()
				return fmt.Errorf("failed to create file %s: %v", tmpSegmentPath, err)
			}

			// Copy the content from the response to the file
			written, err := io.Copy(out, resp.Body)
			if err != nil {
				out.Close()
				resp.Body.Close()
				return fmt.Errorf("failed to write data to file %s: %v", tmpSegmentPath, err)
			}

			out.Close()
			resp.Body.Close()
			segments = append(segments, i)
			totalSize += written
			totalSegments++
			chunkSize += written
		}
		// Move the file to the final location
		for _, i := range segments {
			tmpSegmentPath := filepath.Join(tmpDir, fmt.Sprintf("%d_%d", chunk, i))
			finalSegmentPath := filepath.Join(output, fmt.Sprintf("%d_%d", chunk, i))
			err := os.Rename(tmpSegmentPath, finalSegmentPath)
			if err != nil {
				return fmt.Errorf("failed to move file to final location: %v", err)
			}
		}
		if (j+1)%100 == 0 {
			duration := time.Since(startTime)
			fmt.Printf("Downloaded %d/%d chunks; Time taken for last 100 chunks[%d,%d]: %v; Size: %s\n",
				j+1, len(chunks), chunk-99, chunk, duration, utils.HumanReadableBytes(chunkSize))
			startTime = time.Now()
			chunkSize = 0
		}

	}

	fmt.Printf("Chunks downloaded successfully\nTotal Size: %s\nTotal Segments: %d\n", utils.HumanReadableBytes(totalSize), totalSegments)
	return nil
}

func makeChunksList(output string, lastChunk uint64) ([]uint64, error) {
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
	// Return a list of missing chunks
	return chunks, nil
}

func parseFlags() (string, string, uint64, error) {
	output := flag.String("output", "", "Output string")
	url := flag.String("url", "", "URL string")
	lastChunkStr := flag.String("last_chunk", "", "Last chunk as uint64 string")
	flag.Parse()
	if *output == "" || *url == "" || *lastChunkStr == "" {
		fmt.Println("Usage: --output=<output> --url=<url> --last_chunk=<last_chunk>")
		return "", "", 0, fmt.Errorf("missing required flags")
	}
	lastChunk, err := strconv.ParseUint(*lastChunkStr, 10, 64)
	if err != nil {
		fmt.Println("Error parsing last_chunk:", err)
		return "", "", 0, err
	}
	return *output, *url, lastChunk, nil
}
