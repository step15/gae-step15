package step

import (
	"regexp"
	"strings"
)

var bodyRE = regexp.MustCompile("<body>([^<]+)")

func StripTags(maybeHtml string) string {
	r := maybeHtml
	if matches := bodyRE.FindStringSubmatch(maybeHtml); len(matches) > 1 {
		r = matches[1]
	}
	return strings.TrimSpace(r)
}
