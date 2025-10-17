package server

import (
	"crypto/sha256"
	"fmt"
)

func HashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash[:8])
}
