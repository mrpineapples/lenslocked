package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// HMAC is a wrapper around crypto/hmac that makes it easier
// to use in this codebase.
type HMAC struct {
	hmac hash.Hash
}

// NewHMAC creates and returns a new HMAC (type) object.
func NewHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return HMAC{
		hmac: h,
	}
}

// Hash hashes the input string using HMAC with the secret key
// provided when the HMAC (type) object was created.
func (h HMAC) Hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}
