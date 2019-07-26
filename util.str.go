package session

import (
	"strings"
)

// WrapURIExpression splits CDF by "," and trims leading/trailing space,
// then prepends "^" to the string since we're "regexp" matching
// uri paths with strings put here ;)
//
// Aside from that, we leave input content in tact for use in "regexp".
//
// Remember that by the time the response hits the server, it strips
// content such as "/json/" of its last slash ("/json") so don't forget
// we can use the dollar symbol ("$") when expecting end of URI.
func WrapURIExpression(input string) []string {
	data := strings.Split(input, ",")
	for i, handler := range data {
		str := strings.Trim(handler, " ")
		data[i] = str
	}
	return data
}
