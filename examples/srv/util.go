package main

import (
	"bytes"
	"strings"
)

// WReapLeft makes sure that each text node is trimmed of `separator` and also left-wraps text with the `separator`.
func WReapLeft(separator string, text ...string) string {
	for i, t := range text {
		text[i] = strings.Trim(t, separator)
	}
	return WrapLeft(Wrapper(separator, text...), separator)
}

// WrapLeft puts `wrap` at the beginning of the string if not already present.
func WrapLeft(separator string, text string) string {
	result := strings.TrimLeft(text, separator)
	if strings.Index(result, separator) != 0 {
		result = Cat(separator, result)
	}
	return result
}

// Wrapper concatenates text and wraps it like `Wrap` does with `sep`-arator.
func Wrapper(separator string, text ...string) string {
	return Wrap(separator, strings.Join(text, separator))
}

// Wrap wraps text with `wrap`, written for converting "v" to "/v/".
// see: https://blog.golang.org/strings
func Wrap(wrap string, text string) string {
	result := text
	if strings.Index(result, wrap) != 0 {
		result = Cat(wrap, result)
	}
	if strings.LastIndex(result, wrap) != (len(result) - 1) {
		result = Cat(result, wrap)
	}
	return result
}

// Cat - Concatenate a string by way of writing input to a buffer and
// converting returning its .WriteString() function.
func Cat(pInputString ...string) string {
	var buffer bytes.Buffer
	for _, str := range pInputString {
		buffer.WriteString(str)
	}
	return buffer.String() // fmt.Println(buffer.String())
}
