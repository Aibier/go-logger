package logger

import "regexp"

var (
	// regular expression are thread safe and reusable. Compile and reuse beforehand for better performance
	// improvement: 21749 ns/op -> 2444 ns/op
	patternAuthorization = regexp.MustCompile(`(?i)(Authorization:\s*\w+\s\w{3})[^\r\n]*([^\r\n]{3})`)
	patternPassword      = regexp.MustCompile(`(?i)(password"\s*:\s*".{2})[^"]*(.{1}")`)
)

// SecretMask masquerades the secrets from log.
func SecretMask(b []byte) []byte {
	var masked []byte
	masked = patternAuthorization.ReplaceAll(b, []byte("$1*****$2"))
	masked = patternPassword.ReplaceAll(masked, []byte("$1***$2")) // add ending quote
	return masked
}
