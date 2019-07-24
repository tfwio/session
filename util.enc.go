package session

import "encoding/base64"

// fromBase64e gets base-64 StdEncoding (with error)
func fromBase64e(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

// fromBase64 gets base-64 StdEncoding; ignores error.
func fromBase64(input string) []byte {
	result, _ := base64.StdEncoding.DecodeString(input)
	return result
}

// toBase64 gets base-64 StdEncoding
func toBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// toUBase64 gets base-64 URLEncoding
func toUBase64(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

// fromUBase64 gets base-64 URLEncoding
func fromUBase64(input string) string {
	result, _ := base64.URLEncoding.DecodeString(input)
	return string(result)
}

// bytesToBase64 gets base-64 StdEncoding
func bytesToBase64(input []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
