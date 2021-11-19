// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"strconv"
)

// definition represents a service factory definition with additional metadata.
type definition struct {
	Factory  func(Container) interface{}
	Tags     *itemHash
	AliasOf  *definition
	Priority int16
	Shared   bool
	Private  bool
	Kind     string
}

// newDefinition returns a new definition pointer
func newDefinition(factory func(c Container) interface{}, tags *itemHash) (*definition, error) {
	priorty, err := parseIntegerTag(tagPriority, tags)
	if err != nil {
		return nil, err
	}

	shared, err := parseBoolTag(tagShared, tags)
	if err != nil {
		return nil, err
	}

	private, err := parseBoolTag(tagPrivate, tags)
	if err != nil {
		return nil, err
	}

	kind, err := selectKindTag(tags)
	if err != nil {
		return nil, err
	}

	return &definition{
		Factory:  factory,
		Tags:     tags,
		Priority: priorty,
		Shared:   shared,
		Private:  private,
		Kind:     kind,
	}, nil
}

// parseBoolTag looks for a given tag name in tags and returns the corresponding boolean value.
// It returns an error if tag value can not be parsed.
func parseBoolTag(tagName string, tags *itemHash) (bool, error) {
	if !tags.Has(tagName) {
		return false, nil
	}

	raw := tags.Get(tagName).(string)
	if "" == raw {
		return true, nil
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s tag value '%s' is not a valid boolean", tagName, raw)
	}
	return parsed, nil
}

// parseIntegerTag looks for a given tag name in tags and returns the corresponding int16 value.
// It returns an error if tag value can not be parsed.
func parseIntegerTag(tagName string, tags *itemHash) (int16, error) {
	var i int16 = 0
	if !tags.Has(tagName) {
		return i, nil
	}

	raw := tags.Get(tagName).(string)
	if "" == raw {
		return i, nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("%s tag value '%s' is not a valid number", tagName, raw)
	}

	return int16(parsed), nil
}

// selectKindTag looks for one of the tags representing its kind and returns it. If any of the
// the reserved kind tags is found it returns "factory" as the default value. It returns error
// if more than one reserved kind tag is found.
func selectKindTag(tags *itemHash) (string, error) {
	kindTags := []string{tagFactory, tagValue, tagAlias, tagInject}
	found := 0
	kind := tagFactory
	for _, t := range kindTags {
		if tags.Has(t) {
			kind = t
			found++
		}
	}
	if found > 1 {
		return kind, fmt.Errorf("tag '%s' can't be used simultaneously with %v", kind, kindTags)
	}

	return kind, nil
}
