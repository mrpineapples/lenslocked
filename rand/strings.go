package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

// Bytes will generate n random bytes.
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// String generates a byte slice of size nBytes and returns
// a string that is the base64 URL encoded version of it.
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken is a helper function that generates a remember
// token of a predetermined byte size.
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}
