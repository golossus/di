// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// kindTags are the list of reserved tags that represent valid kinds of service definitions.
var kindTags = []string{TagFactory, TagValue, TagAlias, TagInject}

// keyRex is the regular expression used to parse service keys.
var keyRex = regexp.MustCompile(`^[^#]+|#([^#=]+)|=([^#]+)`)

// parseKey looks for tags in the given key. Tags can be specified using the '#' char as separator. The value for a tag
// can be defined by using the '=' char as separator from the tag name and its value. The real key will be the suffix
// of the given key until the first '#'. This function will trim empty spaces of the found key, tag names and valeus.
// As an Example:
//
// 	" some.suffix #tag1 = 2 # tag2"
//
// Will output:
// 	key  = "some.suffix"
// 	tags = {"tag1": "2", "tag2": ""}
func parseKey(raw string) (key string, tags map[string]string) {
	tags = map[string]string{}

	matches := keyRex.FindAllStringSubmatch(raw, -1)

	for i := 0; i < len(matches); i++ {

		if strings.HasPrefix(matches[i][0], "#") {
			tags[strings.TrimSpace(matches[i][1])] = ""
			continue
		}

		if i == 0 {
			key = strings.TrimSpace(matches[i][0])
			continue
		}

		tags[strings.TrimSpace(matches[i-1][1])] = strings.TrimSpace(matches[i][2])
	}

	return key, tags
}


// definition represents a service factory with required metadata by the container to build
// the service instance and manage its dependencies and behaviour.
type definition struct {
	Factory  func(Container) interface{}
	Tags     map[string]string
	AliasOf  *definition
	Priority int16
	Shared   bool
	Private  bool
	Kind     string
}

// newDefinition returns a new definition pointer
func newDefinition(factory func(c Container) interface{}, tagsList ...map[string]string) (*definition, error) {

	tags := mergeTags(tagsList...)

	priority, err := parseIntegerTag(TagPriority, tags)
	if err != nil {
		return nil, err
	}

	shared, err := parseBoolTag(TagShared, tags)
	if err != nil {
		return nil, err
	}

	private, err := parseBoolTag(TagPrivate, tags)
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
		Priority: priority,
		Shared:   shared,
		Private:  private,
		Kind:     kind,
	}, nil
}

// HasTag returns if current definition has a given tag.
func (d *definition) HasTag(tag string) bool {
	_, ok := d.Tags[tag]
	return ok
}

// GetTagOrDefault returns the value of a given tag or the default value in case definition doesn't have the tag.
func (d *definition) GetTagOrDefault(tag string, def string) string {
	if v, ok := d.Tags[tag]; ok {
		return v
	}

	return def
}

// parseBoolTag looks for a given tag name in tags and returns the corresponding boolean value.
// It returns "true" by default if tag has empty value, but it returns an error if tag value can not be parsed.
func parseBoolTag(tagName string, tags map[string]string) (bool, error) {
	tagValue, ok := tags[tagName]
	if !ok {
		return false, nil
	}

	if "" == tagValue {
		return true, nil
	}

	parsed, err := strconv.ParseBool(tagValue)
	if err != nil {
		return false, fmt.Errorf("%s tag value '%s' is not a valid boolean", tagName, tagValue)
	}
	return parsed, nil
}

// parseIntegerTag looks for a given tag name in tags and returns the corresponding int16 value.
// It returns "0" by default if tag has empty value, or it's not found on tags, but it returns an error if tag
// value can not be parsed as int16.
func parseIntegerTag(tagName string, tags map[string]string) (int16, error) {
	i := int16(0)

	tagValue, ok := tags[tagName]
	if !ok {
		return i, nil
	}

	if "" == tagValue {
		return i, nil
	}

	parsed, err := strconv.ParseInt(tagValue, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("%s tag value '%s' is not a valid number", tagName, tagValue)
	}

	return int16(parsed), nil
}

// selectKindTag looks for one of the tags representing its kind and returns it. If none of
// the reserved kind tags is found it returns TagFactory as the default value. It returns error
// if more than one reserved kind tag is found.
func selectKindTag(tags map[string]string) (string, error) {
	kindTag := TagFactory
	kindCount := 0
	for _, tagName := range kindTags {
		if _, ok := tags[tagName]; ok {
			kindTag = tagName
			kindCount++
		}
	}

	if kindCount > 1 {
		return kindTag, fmt.Errorf("tag '%s' can't be used simultaneously with %v", kindTag, kindTags)
	}

	return kindTag, nil
}

// mergeTags merges the list of tag maps into a single map. Previous entries are preserved from being overwritten.
func mergeTags(tagsList ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, tags := range tagsList {
		for tagName, tagValue := range tags {
			if _, ok := merged[tagName]; !ok {
				merged[tagName] = tagValue
			}
		}
	}

	return merged
}
