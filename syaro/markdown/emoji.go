package markdown

import (
	"code.google.com/p/go.net/html"

	"regexp"
)

var reEmoji *regexp.Regexp = regexp.MustCompile(`:[a-z\d+\-_]+:`)

// process emoji :+1:
func processEmoji(n *html.Node) {
	n.Data = string(reEmoji.ReplaceAllFunc([]byte(n.Data), func(b []byte) []byte {
		emoji, ok := shortcodeReplace[string(b)]
		if ok {
			return emoji
		}
		return b
	}))
}
