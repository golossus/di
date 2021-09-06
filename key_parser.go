package di

import (
	"regexp"
	"strings"
)

const keyPattern = `^[^#]+|#([^#=]+)|=([^#]+)`

type keyParser struct {
	rex *regexp.Regexp
}

func newKeyParser() *keyParser {
	return &keyParser{
		rex: regexp.MustCompile(keyPattern),
	}
}

func (p *keyParser) parse(raw string) (key string, tags *itemHash) {
	tags = newItemHash()
	matches := p.rex.FindAllStringSubmatch(raw, -1)

	for i := 0; i < len(matches); i++ {

		if strings.HasPrefix(matches[i][0], "#") {
			tags.set(strings.TrimSpace(matches[i][1]), "")
			continue
		}

		if i == 0 {
			key = strings.TrimSpace(matches[i][0])
			continue
		}

		tags.set(strings.TrimSpace(matches[i-1][1]), strings.TrimSpace(matches[i][2]))
	}

	return key, tags
}
