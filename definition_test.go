// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func dummyFactory(c Container) interface{} {
	return 1
}

func TestNewDefinition(t *testing.T) {
	parser := newKeyParser()

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
				def, _ := newDefinition(dummyFactory, tags)

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
				def, _ := newDefinition(dummyFactory, tags)

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
				def, _ := newDefinition(dummyFactory, tags)

				assert.Equal(t, data.expected, def.Priority)
			})
		}
	})

	t.Run("kind", func(t *testing.T) {
		sharedData := []struct {
			name     string
			key      string
			expected string
		}{
			{"is factory by default", "", "factory"},
			{"is factory", "#factory", "factory"},
			{"is alias", "#alias", "alias"},
			{"is value", "#value", "value"},
			{"is injectable", "#inject", "inject"},
		}

		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				_, tags := parser.parse(data.key)
				def, _ := newDefinition(dummyFactory, tags)

				assert.Equal(t, data.expected, def.Kind)
			})
		}
	})

	t.Run("creation returns error", func(t *testing.T) {
		sharedData := []struct {
			name  string
			key   string
			error string
		}{
			{"if invalid #priority=abc", "#priority=abc", "priority tag value 'abc' is not a valid number"},
			{"if invalid #private=off", "#private=off", "private tag value 'off' is not a valid boolean"},
			{"if invalid #shared=on", "#shared=on", "shared tag value 'on' is not a valid boolean"},
			{"if overlapping kinds", "#factory #value", "tag 'value' can't be used simultaneously with [factory value alias inject]"},
		}

		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				_, tags := parser.parse(data.key)
				_, err := newDefinition(dummyFactory, tags)
				assert.Equal(t, data.error, err.Error())
			})
		}
	})
}
