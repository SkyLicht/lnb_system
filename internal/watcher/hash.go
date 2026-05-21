package watcher

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(sum.Sum(nil)), nil
}
