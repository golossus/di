// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"reflect"
	"sync"
)

type Container interface {
	Get(key string) interface{}
	GetTaggedBy(tag string, values ...string) []interface{}
	GetParameter(key string) interface{}
}

type container struct {
	builder   *containerBuilder
	instances *itemHash
	sealed    bool
	loading   []string
	lock      *sync.Mutex
}

func (c *container) Get(key string) interface{} {
	def := c.builder.GetDefinition(key)
	if c.sealed && def.Private {
		panic(fmt.Sprintf("service with key '%s' is private and can't be retrieved from the container", key))
	}

	if !def.Shared {
		return c.construct(def, key)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.instances.Has(key) {
		i := c.instances.Get(key)
		return reflect.ValueOf(i).Elem().Interface()
	}

	s := c.construct(def, key)

	c.instances.set(key, &s)

	return s
}

func (c *container) GetTaggedBy(tag string, values ...string) []interface{} {
	keys := c.builder.GetTaggedKeys(tag, values)
	defs := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		defs = append(defs, c.Get(key))
	}

	return defs
}

func (c *container) GetParameter(key string) interface{} {
	return c.builder.GetParameter(key)
}

func (c *container) construct(def *definition, key string) interface{} {
	for i := 0; i < len(c.loading); i++ {
		if c.loading[i] == key {
			msg := "circular reference found while building service '%s' at service '%s'"
			panic(fmt.Sprintf(msg, c.loading[0], c.loading[len(c.loading)-1]))
		}
	}

	u := c.unseal()
	u.loading = append(u.loading, key)

	val := reflect.ValueOf(def.Factory).Call([]reflect.Value{reflect.ValueOf(u)})

	return val[0].Interface()
}

func (c *container) unseal() *container {
	if !c.sealed {
		return c
	}

	return &container{
		builder:   c.builder,
		instances: c.instances,
		sealed:    false,
		loading:   make([]string, 0),
		lock:      &sync.Mutex{},
	}
}
