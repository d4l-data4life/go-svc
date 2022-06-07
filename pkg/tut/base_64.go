package tut

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// RandomBytesArray creates an array of random bytes for usage in tests.
func RandomBytesArray(bytesCount int) []byte {
	result := make([]byte, bytesCount)
	if _, err := rand.Read(result); err != nil {
		panic(fmt.Errorf("could not generate random bytes: %w", err))
	}
	return result
}

func MustDecodeB64(b64String string) []byte {
	b, err := base64.StdEncoding.DecodeString(b64String)
	if err != nil {
		panic(fmt.Errorf("this string is not b64: %s: %w", b64String, err))
	}

	return b
}

func RandomBase64Value(bytesCount int) string {
	return base64.StdEncoding.EncodeToString(RandomBytesArray(bytesCount))
}
