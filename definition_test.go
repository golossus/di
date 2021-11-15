// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefinition(t *testing.T) {
	parser := newKeyParser()
	dummyFactory := func(c Container) interface{} {
		return 1
	}

	t.Run("shared", func(t *testing.T) {
		sharedData := []struct {
			name     string
			key      string
			expected bool
		}{
			{"is true if #shared", "#shared", true},
			{"is true if #shared=", "#shared=", true},
			{"is true if #shared=true", "#shared=true", true},
			{"is true if #shared=1", "#shared=1", true},
			{"is false if #shared=false", "#shared=false", false},
			{"is false if #shared=0", "#shared=false", false},
			{"is false if #shared is not in tags", "", false},
		}

		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				_, tags := parser.parse(data.key)
				def := newDefinition(dummyFactory, tags)

				assert.Equal(t, data.expected, def.Shared)
			})
		}
	})

	t.Run("private", func(t *testing.T) {
		sharedData := []struct {
			name     string
			key      string
			expected bool
		}{
			{"is true if #private", "#private", true},
			{"is true if #private=", "#private=", true},
			{"is true if #private=true", "#private=true", true},
			{"is true if #private=1", "#private=1", true},
			{"is false if #private=false", "#private=false", false},
			{"is false if #private=0", "#private=false", false},
			{"is false if #private is not in tags", "", false},
		}

		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				_, tags := parser.parse(data.key)
				def := newDefinition(dummyFactory, tags)

				assert.Equal(t, data.expected, def.Private)
			})
		}
	})

	t.Run("priority", func(t *testing.T) {
		sharedData := []struct {
			name     string
			key      string
			expected int16
		}{
			{"is 0 if #priority", "#priority", 0},
			{"is 0 if #priority=", "#priority=", 0},
			{"is 0 if #priority=0", "#priority=0", 0},
			{"is 1 if #priority=1", "#priority=1", 1},
			{"is -1 if #priority=1", "#priority=-1", -1},
		}

		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				_, tags := parser.parse(data.key)
				def := newDefinition(dummyFactory, tags)

				assert.Equal(t, data.expected, def.Priority)
			})
		}
	})

	t.Run("creation panics", func(t *testing.T) {
		sharedData := []struct {
			name  string
			key   string
			error string
		}{
			{"if invalid #priority=abc", "#priority=abc", "priority tag value 'abc' is not a valid number"},
			{"if invalid #private=off", "#private=off", "private tag value 'off' is not a valid boolean"},
			{"if invalid #shared=on", "#shared=on", "shared tag value 'on' is not a valid boolean"},
		}

		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				_, tags := parser.parse(data.key)

				assert.PanicsWithValue(t, data.error, func() {
					newDefinition(dummyFactory, tags)
				})
			})
		}
	})
}
