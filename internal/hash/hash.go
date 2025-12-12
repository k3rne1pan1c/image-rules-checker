package hash

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// ComputeSHA256 computes the SHA256 hash of a file
func ComputeSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

