// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is the public container interface, used mainly on service factories.
type Container interface {
	Get(key string) interface{}
	GetTaggedBy(tag string, values ...string) []interface{}
}

type container struct {
	builder   *containerBuilder
	instances *itemHash
	sealed    bool
	loading   []string
	lock      *sync.Mutex
}

// Get will retrieve a service form the container by a given key. It will panic if
// service is not found or if the service has been configured as private.
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

// GetTaggedBy returns all services related to a given tag. If values provided, then
// only the services which match with tag and value will be returned. Services are
// sorted by priority defined with the #priotity tag. If not defined, priority is
// zero. Services with higher priority are returned first.
func (c *container) GetTaggedBy(tag string, values ...string) []interface{} {
	keys := c.builder.GetTaggedKeys(tag, values)
	defs := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		defs = append(defs, c.Get(key))
	}

	return defs
}

// MustBuild builds all the public services once to discover unexpected panic on runtime. If given
// true as parameter, singleton services instances will be preserved. On the contrary, any service
// will be removed to have a fresh container.
func (c *container) MustBuild(dry bool) {
	for k, d := range c.builder.definitions.All() {
		if d.(*definition).Private {
			continue
		}
		_ = c.Get(k)
	}

	if dry {
		c = &container{
			builder:   c.builder,
			instances: newItemHash(),
			sealed:    true,
			loading:   make([]string, 0),
			lock:      c.lock,
		}
	}
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

	u.loading = u.loading[:len(u.loading)-1]

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
