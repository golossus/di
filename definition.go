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
	Factory         func(Container) interface{}
	Tags            *itemHash
	Priority        int16
	AliasOf         *definition
}

func newDefinition(factory func(c Container) interface{}, tags *itemHash ) *definition {
	priority := int16(priorityDefault)
	if tags.Has(tagPriority) {
		raw := tags.Get(tagPriority).(string)
		parsed, err := strconv.ParseInt(raw, 10, 16)
		if err != nil {
			panic(fmt.Sprintf("priority value %s is not a valid number", raw))
		}
		priority = int16(parsed)
	}

	return &definition{
		Factory:  factory,
		Tags:     tags,
		Priority: priority,
	};
}

func (d definition) Shared() bool {
	return d.Tags.Has(tagShared)
}

func (d definition) Private() bool {
	return d.Tags.Has(tagPrivate)
}