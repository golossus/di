// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is the public container interface to retrieve built instances of services, though it is used mainly on
// service factories to build other services' dependencies.
type Container interface {
	Get(key string) interface{}
	GetTaggedBy(tag string, values ...string) []interface{}
}

// container is the result of resolving a containerBuilder instance. It can build and return any service previously
// defined in the mentioned containerBuilder.
type container struct {
	builder   *containerBuilder
	instances map[string]interface{}
	sealed    bool
	loading   []string
	lock      *sync.Mutex
}

// Get will retrieve a service form the container by a given key. It will panic if service is not found or if the
// requested service has been configured as private.
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

	if i, ok := c.instances[key]; ok {
		return reflect.ValueOf(i).Elem().Interface()
	}

	s := c.construct(def, key)

	c.instances[key] = &s

	return s
}

// GetTaggedBy returns all services related to a given tag. If values provided, then only the services which match
// with tag and value will be returned. Services are sorted by priority defined with the #priotity tag. If not defined,
// priority is zero. Services with higher priority are returned first.
func (c *container) GetTaggedBy(tag string, values ...string) []interface{} {
	keys := c.builder.GetTaggedKeys(tag, values)
	defs := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		defs = append(defs, c.Get(key))
	}

	return defs
}

// MustBuild builds all the public services at once to discover unexpected panics on runtime. If given false as parameter,
// singleton services instances will be preserved. On the contrary, a "dry" build will be executed and all built services
// will be removed to have a fresh container.
func (c *container) MustBuild(dry bool) {
	for k, d := range c.builder.definitions {
		if d.Private {
			continue
		}
		_ = c.Get(k)
	}

	if dry {
		c = &container{
			builder:   c.builder,
			instances: make(map[string]interface{}),
			sealed:    true,
			loading:   make([]string, 0),
			lock:      c.lock,
		}
	}
}

// construct builds the service from the given definition. It detects circular referenced dependencies by checking if
// the key has already been built in current dependencies graph.
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

// unseal returns an unsealed version of current container to allow private services to be injected in other services.
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
