package btc

import "encoding/hex"

func BytesToBtcHash(bytes []interface{}) string {
	byteSlice := make([]byte, len(bytes))
	for i, v := range bytes {
		byteSlice[i] = byte(v.(float64))
	}

	// reverse bytes slice
	for i := 0; i < len(byteSlice)/2; i++ {
		j := len(byteSlice) - 1 - i
		byteSlice[i], byteSlice[j] = byteSlice[j], byteSlice[i]
	}

	return hex.EncodeToString(byteSlice)
}
