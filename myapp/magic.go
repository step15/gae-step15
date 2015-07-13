package step

import (
	"regexp"
	"strings"
)

var bodyRE = regexp.MustCompile("(?is:<body>(.+))")
var tagRE = regexp.MustCompile("(?is:<[^>]*>)")

func dropTags(html string) string {
	r := ""
	for _, x := range tagRE.Split(html, -1) {
		r += x + " "
	}
	return r
}

func StripTags(maybeHtml string) string {
	r := maybeHtml
	if matches := bodyRE.FindStringSubmatch(maybeHtml); len(matches) > 1 {
		r = matches[1]
	}
	r = dropTags(r)
	r = strings.Replace(r, "\n", "", -1)
	r = strings.TrimSpace(r)
	return r
}
