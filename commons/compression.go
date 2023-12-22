package commons

import (
	"bytes"
	"compress/zlib"
	"io"
)

func zCompress(data []byte) (compressedData []byte, err error) {
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err = writer.Write(data); err == nil {
		if writer.Close() == nil {
			return buf.Bytes(), nil
		}
	}
	return
}

func zDecompress(compressedData []byte) (data []byte, err error) {
	reader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err == nil {
		defer reader.Close()
		return io.ReadAll(reader)
	}
	return
}

// Compress compresses data with zlib. If the compressed data is larger than the original data, the original data is used. A leading byte is used to indicate whether the data is compressed.
func Compress(data []byte) (compressedBytes []byte, err error) {
	dataSize := len(data)
	compressedData, err := zCompress(data)
	if err == nil {
		if dataSize <= len(compressedData) {
			compressedBytes = append([]byte{0}, data...)
		} else {
			compressedBytes = append([]byte{1}, compressedData...)
		}
	}
	return compressedBytes, err
}

// Decompress decompresses data with zlib. A leading byte is used to indicate whether the data is compressed.
func Decompress(compressedBytes []byte) (data []byte, err error) {
	preProcessedBytes := compressedBytes[1:]
	if compressedBytes[0] == 1 {
		return zDecompress(preProcessedBytes)
	}
	return preProcessedBytes, nil
}
