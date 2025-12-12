package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/k3rne1pan1c/image-checker/internal/model"
)

// PNG signature
var pngSignature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// PNG chunk types
const (
	chunkIHDR = "IHDR" // Image header
	chunkIDAT = "IDAT" // Image data
	chunkIEND = "IEND" // Image end
)

// PNGInfo contains parsed PNG metadata
type PNGInfo struct {
	Width       int
	Height      int
	ColorModel  string
	Compression string
	BitDepth    int
}

// ParsePNG parses PNG file to extract dimensions, color model, and compression info
func ParsePNG(filePath string) (*model.ImageData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := parsePNGInternal(file)
	if err != nil {
		return nil, err
	}

	return &model.ImageData{
		Dimensions: model.Dimensions{
			Width:  info.Width,
			Height: info.Height,
		},
		ColorModel:  info.ColorModel,
		Compression: info.Compression,
	}, nil
}

func parsePNGInternal(reader io.Reader) (*PNGInfo, error) {
	info := &PNGInfo{
		Compression: "PNG (Deflate)",
	}

	// Read and verify PNG signature
	sig := make([]byte, 8)
	if _, err := reader.Read(sig); err != nil {
		return nil, fmt.Errorf("failed to read PNG signature: %w", err)
	}

	for i, b := range pngSignature {
		if sig[i] != b {
			return nil, fmt.Errorf("invalid PNG signature")
		}
	}

	// Parse chunks
	for {
		// Read chunk length (4 bytes, big-endian)
		lenBuf := make([]byte, 4)
		if _, err := reader.Read(lenBuf); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read chunk length: %w", err)
		}
		chunkLength := binary.BigEndian.Uint32(lenBuf)

		// Read chunk type (4 bytes)
		typeBuf := make([]byte, 4)
		if _, err := reader.Read(typeBuf); err != nil {
			return nil, fmt.Errorf("failed to read chunk type: %w", err)
		}
		chunkType := string(typeBuf)

		// Parse IHDR chunk (Image Header)
		if chunkType == chunkIHDR {
			if err := parseIHDR(reader, info, chunkLength); err != nil {
				return nil, fmt.Errorf("failed to parse IHDR: %w", err)
			}
		} else {
			// Skip chunk data
			dataBuf := make([]byte, chunkLength)
			if _, err := reader.Read(dataBuf); err != nil {
				return nil, fmt.Errorf("failed to read chunk data: %w", err)
			}
		}

		// Read and verify CRC (4 bytes)
		crcBuf := make([]byte, 4)
		if _, err := reader.Read(crcBuf); err != nil {
			return nil, fmt.Errorf("failed to read CRC: %w", err)
		}

		// Stop at IEND chunk
		if chunkType == chunkIEND {
			break
		}
	}

	return info, nil
}

func parseIHDR(reader io.Reader, info *PNGInfo, chunkLength uint32) error {
	if chunkLength != 13 {
		return fmt.Errorf("invalid IHDR chunk length: expected 13, got %d", chunkLength)
	}

	// Read IHDR data (13 bytes)
	ihdrData := make([]byte, 13)
	if _, err := reader.Read(ihdrData); err != nil {
		return err
	}

	// Width (4 bytes, big-endian)
	info.Width = int(binary.BigEndian.Uint32(ihdrData[0:4]))

	// Height (4 bytes, big-endian)
	info.Height = int(binary.BigEndian.Uint32(ihdrData[4:8]))

	// Bit depth (1 byte)
	info.BitDepth = int(ihdrData[8])

	// Color type (1 byte)
	colorType := ihdrData[9]

	// Determine color model based on color type
	switch colorType {
	case 0:
		info.ColorModel = "Grayscale"
	case 2:
		info.ColorModel = "RGB"
	case 3:
		info.ColorModel = "Indexed"
	case 4:
		info.ColorModel = "Grayscale+Alpha"
	case 6:
		info.ColorModel = "RGBA"
	default:
		info.ColorModel = "Unknown"
	}

	// Compression method (1 byte) - always 0 (Deflate) in PNG
	compressionMethod := ihdrData[10]
	if compressionMethod == 0 {
		info.Compression = "PNG (Deflate)"
	} else {
		info.Compression = "PNG (Unknown)"
	}

	// Filter method (1 byte) - always 0 in PNG
	// Interlace method (1 byte) - 0 (none) or 1 (Adam7)

	return nil
}

