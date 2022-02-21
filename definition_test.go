// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func dummyFactory(_ Container) interface{} {
	return 1
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		raw  string
		key  string
		tags map[string]string
	}{
		{"", "", map[string]string{}},
		{"Key", "Key", map[string]string{}},
		{" Key ", "Key", map[string]string{}},
		{" Key #tag1", "Key", map[string]string{"tag1": ""}},
		{" Key #tag1 #tag2", "Key", map[string]string{"tag1": "", "tag2": ""}},
		{" Key #tag1=1a #tag2=2b ", "Key", map[string]string{"tag1": "1a", "tag2": "2b"}},
		{" Key #tag1=1a #tag2=2b cc", "Key", map[string]string{"tag1": "1a", "tag2": "2b cc"}},
		{" Key #tag1=1a #tag2=2b cc ", "Key", map[string]string{"tag1": "1a", "tag2": "2b cc"}},
		{" Key #tag1 #tag2 3 4 ", "Key", map[string]string{"tag1": "", "tag2 3 4": ""}},
		{" Key #tag1 =1a #tag2 = 2b ", "Key", map[string]string{"tag1": "1a", "tag2": "2b"}},
		{" Key #tag1 = #tag2 =", "Key", map[string]string{"tag1": "", "tag2": ""}},
		{" some.suffix #tag1 = 2 # tag2", "some.suffix", map[string]string{"tag1": "2", "tag2": ""}},
		{"#tag1", "", map[string]string{"tag1": ""}},
		{"=tag1", "=tag1", map[string]string{}},
	}

	for _, data := range tests {
		t.Run(data.raw, func(t *testing.T) {
			key, tags := parseKey(data.raw)
			assert.Equal(t, data.key, key)
			assert.Equal(t, data.tags, tags)
		})
	}
}

func TestParseBoolTag(t *testing.T) {
	testData := []struct {
		test     string
		tags     map[string]string
		expected bool
	}{
		{"is true if tag with empty value", map[string]string{"tag": ""}, true},
		{"is true if tag value is true", map[string]string{"tag": "true"}, true},
		{"is true if tag value is 1", map[string]string{"tag": "1"}, true},
		{"is false if tag value is false", map[string]string{"tag": "false"}, false},
		{"is false if tag value is 0", map[string]string{"tag": "0"}, false},
		{"is false if tag is not present", map[string]string{}, false},
	}

	for _, data := range testData {
		t.Run(data.test, func(t *testing.T) {
			b, _ := parseBoolTag("tag", data.tags)
			assert.Equal(t, data.expected, b)
		})
	}

	t.Run("fails if tag value not boolean", func(t *testing.T) {
		b, err := parseBoolTag("tag", map[string]string{"tag": "dummy"})
		assert.Equal(t, false, b)
		assert.NotNil(t, err)
	})
}

func TestParseIntegerTag(t *testing.T) {
	testData := []struct {
		test     string
		tags     map[string]string
		expected int16
	}{
		{"is 0 if tag is not present", map[string]string{}, 0},
		{"is 0 if tag with empty value", map[string]string{"tag": ""}, 0},
		{"is 0 if tag value is 0", map[string]string{"tag": "0"}, 0},
		{"is 1 if tag value is 1", map[string]string{"tag": "1"}, 1},
		{"is -1 if tag value is -1", map[string]string{"tag": "-1"}, -1},
	}

	for _, data := range testData {
		t.Run(data.test, func(t *testing.T) {
			b, _ := parseIntegerTag("tag", data.tags)
			assert.Equal(t, data.expected, b)
		})
	}

	t.Run("fails if tag value not int16", func(t *testing.T) {
		b, err := parseIntegerTag("tag", map[string]string{"tag": "dummy"})
		assert.Equal(t, int16(0), b)
		assert.NotNil(t, err)
	})
}

func TestSelectKindTag(t *testing.T) {
	for _, kindTag := range kindTags {
		t.Run(fmt.Sprintf("returns %s as kind if present", kindTag), func(t *testing.T) {
			b, err := selectKindTag(map[string]string{"tag1": "", kindTag: "", "tag2": ""})
			assert.Equal(t, kindTag, b)
			assert.Nil(t, err)
		})
	}

	t.Run("returns factory if any kind tag present", func(t *testing.T) {
		b, err := selectKindTag(map[string]string{"tag1": ""})
		assert.Equal(t, TagFactory, b)
		assert.Nil(t, err)
	})

	t.Run("returns error if multiple kind tags present", func(t *testing.T) {
		b, err := selectKindTag(map[string]string{TagFactory: "", TagValue: ""})
		assert.Equal(t, TagValue, b)
		assert.NotNil(t, err)
	})
}

func TestMergeTags(t *testing.T) {
	t.Run("merges all tags and preserves the former keys", func(t *testing.T) {
		a := map[string]string{"tag1": "preserved"}
		b := map[string]string{"tag2": ""}
		c := map[string]string{"tag3": "any"}

		expected := map[string]string{"tag1": "preserved", "tag2": "", "tag3": "any"}

		merged := mergeTags(a, b, c)
		assert.Equal(t, expected, merged)
	})
}

func TestNewDefinition(t *testing.T) {
	t.Run("is created with empty tags", func(t *testing.T) {
		def, _ := newDefinition(dummyFactory)
		assert.Equal(t, false, def.Shared)
		assert.Equal(t, false, def.Private)
		assert.Equal(t, int16(0), def.Priority)
		assert.Equal(t, TagFactory, def.Kind)
		assert.Equal(t, map[string]string{}, def.Tags)
		assert.Equal(t, dummyFactory(nil), def.Factory(nil))
	})

	t.Run("is created with custom tags", func(t *testing.T) {
		custom := map[string]string{
			TagValue:    "",
			TagPrivate:  "",
			TagShared:   "1",
			TagPriority: "9",
		}
		def, _ := newDefinition(dummyFactory, custom)

		assert.Equal(t, true, def.Shared)
		assert.Equal(t, true, def.Private)
		assert.Equal(t, int16(9), def.Priority)
		assert.Equal(t, TagValue, def.Kind)
		assert.Equal(t, custom, def.Tags)
		assert.Equal(t, dummyFactory(nil), def.Factory(nil))
	})

	t.Run("creation returns error", func(t *testing.T) {
		testData := []struct {
			name  string
			tags  map[string]string
			error string
		}{
			{"if invalid priority value", map[string]string{TagPriority: "abc"}, "priority tag value 'abc' is not a valid number"},
			{"if invalid private value", map[string]string{TagPrivate: "off"}, "private tag value 'off' is not a valid boolean"},
			{"if invalid shared value", map[string]string{TagShared: "on"}, "shared tag value 'on' is not a valid boolean"},
			{"if multiple kind tags", map[string]string{TagFactory: "", TagValue: ""}, "tag 'value' can't be used simultaneously with [factory value alias inject]"},
		}

		for _, data := range testData {
			t.Run(data.name, func(t *testing.T) {
				_, err := newDefinition(dummyFactory, data.tags)
				assert.Equal(t, data.error, err.Error())
			})
		}
	})
}

func TestDefinition_HasTag(t *testing.T) {
	def, _ := newDefinition(dummyFactory, map[string]string{"exists": "abc"})

	t.Run("returns true if tag exists", func(t *testing.T) {
		assert.True(t, def.HasTag("exists"))
	})

	t.Run("returns false if tag does not exist", func(t *testing.T) {
		assert.False(t, def.HasTag("not-exists"))
	})
}

func TestDefinition_GetTag(t *testing.T) {
	def, _ := newDefinition(dummyFactory, map[string]string{"exists": "abc"})

	t.Run("returns tag value if exists", func(t *testing.T) {
		assert.Equal(t, "abc",  def.GetTag("exists"))
	})

	t.Run("returns alternative if tag does not exist", func(t *testing.T) {
		assert.Equal(t, "alternative",  def.GetTag("not-exists", "alternative"))
	})

	t.Run("returns empty string if tag does not exist and alternative not given", func(t *testing.T) {
		assert.Equal(t, "",  def.GetTag("not-exists"))
	})
}
