// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

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
		{"Key", "Key", map[string]interface{}{}},
		{" Key ", "Key", map[string]interface{}{}},
		{" Key #tag1", "Key", map[string]interface{}{"tag1": ""}},
		{" Key #tag1 #tag2", "Key", map[string]interface{}{"tag1": "", "tag2": ""}},
		{" Key #tag1=1a #tag2=2b ", "Key", map[string]interface{}{"tag1": "1a", "tag2": "2b"}},
		{" Key #tag1=1a #tag2=2b cc", "Key", map[string]interface{}{"tag1": "1a", "tag2": "2b cc"}},
		{" Key #tag1=1a #tag2=2b cc ", "Key", map[string]interface{}{"tag1": "1a", "tag2": "2b cc"}},
		{" Key #tag1 #tag2 3 4 ", "Key", map[string]interface{}{"tag1": "", "tag2 3 4": ""}},
		{" Key #tag1 =1a #tag2 = 2b ", "Key", map[string]interface{}{"tag1": "1a", "tag2": "2b"}},
		{" Key #tag1 = #tag2 =", "Key", map[string]interface{}{"tag1": "", "tag2": ""}},
		{"#tag1", "", map[string]interface{}{"tag1": ""}},
		{"=tag1", "=tag1", map[string]interface{}{}},
	}

	for _, data := range tests {
		parser := newKeyParser()
		t.Run(data.raw, func(t *testing.T) {
			key, tags := parser.parse(data.raw)
			assert.Equal(t, data.key, key)
			assert.Equal(t, data.tags, tags.All())
		})
	}
}
