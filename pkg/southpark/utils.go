package southpark

import (
	"strings"
)

// Manual implementation of go1.20's CutPrefix
// to support older go versions
func cutPrefix(s, prefix string) (after string, found bool) {
	if !strings.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}
