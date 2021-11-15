// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"strconv"
)

//definition represents a service factory definition with additional metadata.
type definition struct {
	Factory  func(Container) interface{}
	Tags     *itemHash
	AliasOf  *definition
	Priority int16
	Shared   bool
	Private  bool
}

//newDefinition returns a new definition pointer
func newDefinition(factory func(c Container) interface{}, tags *itemHash) *definition {
	return &definition{
		Factory:  factory,
		Tags:     tags,
		Priority: parseIntegerTag(tagPriority, tags),
		Shared:   parseBoolTag(tagShared, tags),
		Private:  parseBoolTag(tagPrivate, tags),
	}
}

//parseBoolTag looks for a given tag name in tags and returns the corresponding boolean value
func parseBoolTag(tagName string, tags *itemHash) bool {
	if !tags.Has(tagName) {
		return false
	}

	raw := tags.Get(tagName).(string)
	if "" == raw {
		return true
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		panic(fmt.Sprintf("%s tag value '%s' is not a valid boolean", tagName, raw))
	}
	return parsed
}

//parseIntegerTag looks for a given tag name in tags and returns the corresponding int16 value
func parseIntegerTag(tagName string, tags *itemHash) int16 {
	var i int16 = 0
	if !tags.Has(tagName) {
		return i
	}

	raw := tags.Get(tagName).(string)
	if "" == raw {
		return i
	}

	parsed, err := strconv.ParseInt(raw, 10, 16)
	if err != nil {
		panic(fmt.Sprintf("%s tag value '%s' is not a valid number", tagName, raw))
	}

	return int16(parsed)
}
