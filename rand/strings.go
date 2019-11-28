package rand

import (
	"crypto/rand"
	"encoding/base64"
)

// Size of remember token in bytes.
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

// NBytes returns the number of bytes used in the base64 URL encoded string.
func NBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return -1, err
	}
	return len(b), nil
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
