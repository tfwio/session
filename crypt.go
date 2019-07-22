package session

import (
	"crypto/rand"
	"runtime"

	"golang.org/x/crypto/argon2"
)

var (
	defaultHashMem    = uint32(64 * 1024)
	defaultHashTime   = uint32(2)
	defaultHashKeyLen = uint32(32)
	defaultHashThread = int32(runtime.NumCPU())
)

// Override allows you to override default hash creation settings.
// Set a value to -1 to persist default(s).
// *note*: that defaults are `uint32`.
func Override(hashMemSize int64, hashTime int64, hashKeyLength int64) {
	if hashMemSize != -1 {
		defaultHashMem = uint32(hashMemSize)
	}
	if hashTime != -1 {
		defaultHashTime = uint32(hashTime)
	}
	if hashKeyLength != -1 {
		defaultHashKeyLen = uint32(hashKeyLength)
	}
}

// NewSaltCSRNG CSRNG salt
func NewSaltCSRNG(c int) []byte {
	b := make([]byte, c)
	rand.Read(b)
	return b
}

// NewSaltString calls NewSaltCSRNG and converts the result to base64 string.
func NewSaltString(c int) string {
	return bytesToBase64(NewSaltCSRNG(c))
}

// copyTo copys bytes into a byte array.
func copyTo(dst []byte, src []byte, offset int) {
	for j, k := range src {
		dst[offset+j] = k
	}
}

// compareBytes returns true on match.
// It compares length and each byte of inputs.
func compareBytes(a []byte, b []byte) bool {

	if len(a) != len(b) {
		return false
	}
	for i, j := range a {
		if b[i] != j {
			return false
		}
	}
	return true
}

// GetHash dammit.
func GetHash(pass []byte, salt []byte) []byte {

	salty := make([]byte, len(salt)+len(pass))
	copyTo(salty, salt, 0)         // add salt
	copyTo(salty, pass, len(salt)) // add username to end

	// 3 passes(time) 32 mB
	return argon2.IDKey(pass, salty, defaultHashTime, defaultHashMem, uint8(runtime.NumCPU()), defaultHashKeyLen)
}

// GetPasswordHash makes a hash from password and salt.
func GetPasswordHash(password string, salt []byte) []byte {
	return GetHash([]byte(password), salt)
}

// CheckPassword compares salt/password against an existing hash.
func CheckPassword(password string, salt []byte, hash []byte) bool {
	if compareBytes(GetPasswordHash(password, salt), hash) {
		return true
	}
	return false
}
