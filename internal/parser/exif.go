package parser

import (
	"fmt"
	"os"
	"time"

	"github.com/k3rne1pan1c/image-checker/internal/model"
	"github.com/rwcarlsen/goexif/exif"
)

// ParseEXIF extracts EXIF metadata from an image file
func ParseEXIF(filePath string) (*model.EXIFData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		// EXIF may not be present, return nil without error
		return nil, nil
	}

	exifData := &model.EXIFData{}

	// Extract camera maker
	if maker, err := x.Get(exif.Make); err == nil {
		if makerStr, err := maker.StringVal(); err == nil {
			exifData.Maker = makerStr
		}
	}

	// Extract timestamp
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		if dateTimeStr, err := dateTime.StringVal(); err == nil {
			// Parse EXIF datetime format: "2006:01:02 15:04:05"
			if t, err := time.Parse("2006:01:02 15:04:05", dateTimeStr); err == nil {
				exifData.Timestamp = t.Format(time.RFC3339)
			}
		}
	}

	// Extract GPS coordinates
	if lat, lon, err := x.LatLong(); err == nil {
		exifData.GPS = &model.GPSData{
			Lat: lat,
			Lon: lon,
		}
	}

	// Extract orientation
	if orient, err := x.Get(exif.Orientation); err == nil {
		if orientInt, err := orient.Int(0); err == nil {
			exifData.Orientation = int(orientInt)
		}
	}

	return exifData, nil
}

// HasEXIF checks if a file has EXIF data without fully parsing it
func HasEXIF(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	_, err = exif.Decode(file)
	return err == nil
}

