package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/k3rne1pan1c/image-checker/internal/model"
)

// JPEG markers
const (
	markerSOI  = 0xFFD8 // Start of Image
	markerSOF0 = 0xFFC0 // Start of Frame (Baseline DCT)
	markerSOF1 = 0xFFC1 // Start of Frame (Extended Sequential DCT)
	markerSOF2 = 0xFFC2 // Start of Frame (Progressive DCT)
	markerDHT  = 0xFFC4 // Define Huffman Table
	markerDQT  = 0xFFDB // Define Quantization Table
	markerDRI  = 0xFFDD // Define Restart Interval
	markerSOS  = 0xFFDA // Start of Scan (entropy-coded data)
	markerAPP0 = 0xFFE0 // Application Data
	markerAPP1 = 0xFFE1 // Application Data (often EXIF)
	markerCOM  = 0xFFFE // Comment
	markerEOI  = 0xFFD9 // End of Image
)

// JPEGInfo contains parsed JPEG metadata
type JPEGInfo struct {
	Width       int
	Height      int
	ColorModel  string
	Compression string
	HasDHT      bool // Has Huffman tables (entropy coding)
	HasDQT      bool // Has Quantization tables
}

// ParseJPEG parses JPEG file to extract dimensions, color model, and compression info
func ParseJPEG(filePath string) (*model.ImageData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := parseJPEGInternal(file)
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

func parseJPEGInternal(reader io.Reader) (*JPEGInfo, error) {
	info := &JPEGInfo{
		Compression: "JPEG",
	}

	buf := make([]byte, 2)
	
	// Read SOI marker
	if _, err := reader.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to read SOI marker: %w", err)
	}
	if binary.BigEndian.Uint16(buf) != markerSOI {
		return nil, fmt.Errorf("invalid JPEG: missing SOI marker")
	}

	// Parse segments
	for {
		// Read marker
		if _, err := reader.Read(buf); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read marker: %w", err)
		}

		marker := binary.BigEndian.Uint16(buf)

		// Skip padding bytes (0xFF may be followed by 0x00)
		for marker == 0xFFFF {
			if _, err := reader.Read(buf); err != nil {
				return nil, fmt.Errorf("failed to skip padding: %w", err)
			}
			marker = binary.BigEndian.Uint16(buf)
		}

		// Handle different segment types
		switch marker {
		case markerSOF0, markerSOF1, markerSOF2:
			// Start of Frame - contains dimensions and color info
			if err := parseSOF(reader, info, marker); err != nil {
				return nil, fmt.Errorf("failed to parse SOF: %w", err)
			}
			// Once we have dimensions, we can return early if we don't need more
			if info.Width > 0 && info.Height > 0 {
				// Continue to check for entropy segments
			}

		case markerDHT:
			// Define Huffman Table - entropy coding
			info.HasDHT = true
			if err := skipSegment(reader); err != nil {
				return nil, fmt.Errorf("failed to skip DHT: %w", err)
			}

		case markerDQT:
			// Define Quantization Table - compression quality indicator
			info.HasDQT = true
			if err := skipSegment(reader); err != nil {
				return nil, fmt.Errorf("failed to skip DQT: %w", err)
			}

		case markerSOS:
			// Start of Scan - marks beginning of entropy-coded data
			// This is where the actual compressed image data starts
			if err := skipSegment(reader); err != nil {
				return nil, fmt.Errorf("failed to skip SOS: %w", err)
			}
			// After SOS, we're in entropy-coded data, skip until next marker
			if err := skipEntropyData(reader); err != nil {
				return nil, fmt.Errorf("failed to skip entropy data: %w", err)
			}

		case markerEOI:
			// End of Image
			break

		case markerAPP0, markerAPP1, markerCOM, markerDRI:
			// Application data, comments, restart interval - skip
			if err := skipSegment(reader); err != nil {
				return nil, fmt.Errorf("failed to skip segment: %w", err)
			}

		default:
			// Unknown marker - try to skip if it's a segment with length
			if marker >= 0xFFE0 && marker <= 0xFFEF {
				// APP markers
				if err := skipSegment(reader); err != nil {
					return nil, fmt.Errorf("failed to skip APP segment: %w", err)
				}
			} else if marker == 0xFF00 {
				// Padding byte, continue
				continue
			} else {
				// Unknown marker, try to continue
				if err := skipSegment(reader); err != nil {
					// If we can't skip, we might be at the end
					break
				}
			}
		}

		if marker == markerEOI {
			break
		}
	}

	return info, nil
}

func parseSOF(reader io.Reader, info *JPEGInfo, marker uint16) error {
	// Read segment length (2 bytes)
	lenBuf := make([]byte, 2)
	if _, err := reader.Read(lenBuf); err != nil {
		return err
	}
	length := binary.BigEndian.Uint16(lenBuf)

	// Read precision (1 byte)
	precision := make([]byte, 1)
	if _, err := reader.Read(precision); err != nil {
		return err
	}

	// Read height (2 bytes)
	heightBuf := make([]byte, 2)
	if _, err := reader.Read(heightBuf); err != nil {
		return err
	}
	info.Height = int(binary.BigEndian.Uint16(heightBuf))

	// Read width (2 bytes)
	widthBuf := make([]byte, 2)
	if _, err := reader.Read(widthBuf); err != nil {
		return err
	}
	info.Width = int(binary.BigEndian.Uint16(widthBuf))

	// Read number of components (1 byte)
	componentsBuf := make([]byte, 1)
	if _, err := reader.Read(componentsBuf); err != nil {
		return err
	}
	numComponents := int(componentsBuf[0])

	// Determine color model based on number of components
	switch numComponents {
	case 1:
		info.ColorModel = "Grayscale"
	case 3:
		info.ColorModel = "RGB"
	case 4:
		info.ColorModel = "CMYK"
	default:
		info.ColorModel = "Unknown"
	}

	// Determine compression type based on SOF marker
	switch marker {
	case markerSOF0:
		info.Compression = "JPEG (Baseline DCT)"
	case markerSOF1:
		info.Compression = "JPEG (Extended Sequential DCT)"
	case markerSOF2:
		info.Compression = "JPEG (Progressive DCT)"
	}

	// Skip remaining component data (3 bytes per component)
	remaining := int(length) - 8 // Already read: length(2) + precision(1) + height(2) + width(2) + components(1)
	skipBuf := make([]byte, remaining)
	if _, err := reader.Read(skipBuf); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func skipSegment(reader io.Reader) error {
	lenBuf := make([]byte, 2)
	if _, err := reader.Read(lenBuf); err != nil {
		return err
	}
	length := binary.BigEndian.Uint16(lenBuf)
	
	// Skip segment data (length includes the 2 bytes for length itself)
	skipBuf := make([]byte, int(length)-2)
	if _, err := reader.Read(skipBuf); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func skipEntropyData(reader io.Reader) error {
	// Entropy-coded data continues until we find a marker (0xFF followed by non-zero byte)
	buf := make([]byte, 1)
	prevWasFF := false

	for {
		if _, err := reader.Read(buf); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if buf[0] == 0xFF {
			prevWasFF = true
		} else if prevWasFF {
			// Found a marker, back up one byte
			// We need to seek back, but Reader doesn't support Seek
			// For now, we'll just break and let the caller handle it
			break
		} else {
			prevWasFF = false
		}
	}

	return nil
}

