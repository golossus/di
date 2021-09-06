package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testData struct {
	raw  string
	key  string
	tags map[string]interface{}
}

func TestKeyParser_Parse(t *testing.T) {

	tests := []testData{
		{"key", "key", map[string]interface{}{}},
		{" key ", "key", map[string]interface{}{}},
		{" key #tag1", "key", map[string]interface{}{"tag1": ""}},
		{" key #tag1 #tag2", "key", map[string]interface{}{"tag1": "", "tag2": ""}},
		{" key #tag1=1a #tag2=2b ", "key", map[string]interface{}{"tag1": "1a", "tag2": "2b"}},
		{" key #tag1=1a #tag2=2b cc", "key", map[string]interface{}{"tag1": "1a", "tag2": "2b cc"}},
		{" key #tag1=1a #tag2=2b cc ", "key", map[string]interface{}{"tag1": "1a", "tag2": "2b cc"}},
		{" key #tag1 #tag2 3 4 ", "key", map[string]interface{}{"tag1": "", "tag2 3 4": ""}},
		{" key #tag1 =1a #tag2 = 2b ", "key", map[string]interface{}{"tag1": "1a", "tag2": "2b"}},
		{" key #tag1 = #tag2 =", "key", map[string]interface{}{"tag1": "", "tag2": ""}},
		{"#tag1", "", map[string]interface{}{"tag1": ""}},
		{"=tag1", "=tag1", map[string]interface{}{}},
	}

	for _, data := range tests {
		parser := newKeyParser()
		t.Run(data.raw, func(t *testing.T) {
			key, tags := parser.parse(data.raw)
			assert.Equal(t, data.key, key)
			assert.Equal(t, data.tags, tags.all())
		})
	}
}
