// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

const (
	tagShared   = "shared"
	tagPrivate  = "private"
	tagPriority = "priority"
	tagInject   = "inject"
	tagValue    = "value"
	tagAlias    = "alias"
	tagFactory  = "factory"
)

type Binding struct {
	Key    string
	Target interface{}
}

// Provider allows to provide definitions into containerBuilder. Binding dependencies
// might not be available yet during the call to this method.
type Provider interface {
	Provide(builder ContainerBuilder)
}

// ProviderFunc adapts a normal func into a Provider.
type ProviderFunc func(ContainerBuilder)

func (f ProviderFunc) Provide(b ContainerBuilder) {
	f(b)
}

// Resolver allows to resolve definitions into containerBuilder once All services
// definitions are available.
type Resolver interface {
	Resolve(builder ContainerBuilder)
}

// ResolverFunc adapts a normal func into a Resolver.
type ResolverFunc func(ContainerBuilder)

func (f ResolverFunc) Resolve(b ContainerBuilder) {
	f(b)
}

// ContainerBuilder interface declares the public api for containerBuilder type.
type ContainerBuilder interface {
	SetAll(all ...Binding)
	SetValue(key string, value interface{}) *definition
	SetFactory(key string, factory interface{}) *definition
	SetInjectable(key string, value interface{}) *definition
	SetAlias(key, def string) *definition
	HasDefinition(key string) bool
	GetDefinition(key string) *definition
	GetTaggedKeys(tag string, values []string) []string
}

// containerBuilder implements ContainerBuilder interface to bind service definitions
// and resolve the final service container.
type containerBuilder struct {
	definitions *itemHash
	parser      *keyParser
	providers   []Provider
	resolvers   []Resolver
	resolved    bool
	lock        *sync.Mutex
}

// NewContainerBuilder returns a pointer to a new containerBuilder instance.
func NewContainerBuilder() *containerBuilder {
	return &containerBuilder{
		definitions: newItemHash(),
		parser:      newKeyParser(),
		providers:   make([]Provider, 0),
		resolvers:   make([]Resolver, 0),
		resolved:    false,
		lock:        &sync.Mutex{},
	}
}

// SetValue adds a new value or instance to the container on a given Key.
func (c *containerBuilder) SetValue(key string, value interface{}) *definition {
	return c.SetFactory(key, func(c Container) interface{} {
		return value
	})
}

// SetAll sets given bindings into the containerBuilder. Reserved tags: #value, #alias,
// #inject and #factory; are used to set the correct service definition. If any of the
// reserved tags is indicated, then #factory will be considered as default. Reserved
// tags are all mutually exclusive and adding more than one at a time will panic.
//
//   b.SetAll([]Binding{
//		{Key: "key1 #factory", Target: func(c Container) interface{} {
//			return 1
//		}},
//		{Key: "key2 #value", Target: 2},
//		{Key: "key3 #alias", Target: "key2"},
//		{Key: "key4 #inject", Target: struct{}{}},
//		{Key: "key5", Target: func(c Container) interface{} {  	// <- defaults to #factory
//			return 5
//		}},
//		{Key: "key4 #value #alias", Target: "key2}, 			// <- will panic
//	}...)
func (c *containerBuilder) SetAll(all ...Binding) {
	for _, i := range all {
		_, tags := c.parser.parse(i.Key)
		switch {
		case tags.Has(tagAlias):
			c.SetAlias(i.Key, i.Target.(string))
		case tags.Has(tagValue):
			c.SetValue(i.Key, i.Target)
		case tags.Has(tagInject):
			c.SetInjectable(i.Key, i.Target)
		case tags.Has(tagFactory):
			fallthrough
		default:
			c.SetFactory(i.Key, i.Target)
		}
	}
}

// SetInjectable adds a new definition for a given struct to the container referenced
// by a given Key. Given struct must contain at least one public member labeled with
// the "inject" label:
//	 type SomeType struct {
//	 	 field string `inject:"service.to.inject.key"`
//	 }
//
// Keys can contain tags in the form of "#tag[=value]" where the value
// part can be omitted.
//
// Reserved tags "#shared" and "#private" can be used to make a
// definition shared (factory will return a singleton) and private (service will
// be available to be injected as dependency but not available to be retrieved
// from current container).
func (c *containerBuilder) SetInjectable(key string, i interface{}) *definition {
	t := reflect.TypeOf(i)
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("invalid injectable for key %s, only structs can be injectables", key))
	}

	fields := make(map[int]string)
	for j := 0; j < t.NumField(); j++ {
		f := t.Field(j)
		k, ok := f.Tag.Lookup("inject")
		if !ok {
			continue
		}

		if len(f.PkgPath) != 0 {
			panic(fmt.Sprintf("unexported field %s/%s can not be injected", f.PkgPath, f.Name))
		}

		if len(k) == 0 {
			panic(fmt.Sprintf("no injection key present for field %s: %s", t.Name(), f.Name))
		}

		fields[j] = k
	}

	d := c.SetFactory(key, func(c Container) interface{} {
		t := reflect.New(t)
		e := t.Elem()
		for i, k := range fields {
			p := c.Get(k)
			v := reflect.ValueOf(p)
			e.Field(i).Set(v)
		}

		if isPtr {
			return t.Interface()
		}

		return e.Interface()
	})

	return d

}

// SetFactory adds a new definition to the container referenced by a given
// Key. Keys can contain tags in the form of "#tag[=value]" where the value part
// can be omitted.
//
// Reserved tags "#shared" and "#private" can be used to make a definition shared
// (factory will return a singleton) and private (service will be available to be
// injected as dependency but not available to be retrieved from current container).
func (c *containerBuilder) SetFactory(key string, factory interface{}) *definition {
	c.panicIfResolved()

	f, ok := factory.(func(c Container) interface{})
	if !ok {
		panic(fmt.Sprintf("type '%T' for key '%s' is not a valid factory", factory, key))
	}

	k, tags := c.parser.parse(key)

	d, err := newDefinition(f, tags)
	if err != nil {
		panic(fmt.Sprintf("%s for key '%s'", err, key))
	}
	c.definitions.set(k, d)

	return d
}

// HasDefinition returns true if definition for the Key exists in the container.
func (c *containerBuilder) HasDefinition(key string) bool {
	return c.definitions.Has(key)
}

// GetDefinition retrieves a container definition for the Key or panics if not found.
func (c *containerBuilder) GetDefinition(key string) *definition {
	return c.definitions.Get(key).(*definition)
}

// SetAlias sets an alias for an existing definition. Aliases inherit the aliased service
// factory but the can have thei own set of tags. As an example, a service might be "private"
// and the corresponding alias can be public or even a singleton.
func (c *containerBuilder) SetAlias(key, def string) *definition {
	k, _ := c.parser.parse(key)

	if !c.definitions.Has(def) {
		panic(fmt.Sprintf("definition with id '%s' does not exist and alias cannot be set", def))
	}

	if c.definitions.Has(k) && c.definitions.Get(k).(*definition).AliasOf == nil {
		panic(fmt.Sprintf("definition with id '%s' already exists and alias cannot be set", key))
	}

	aliased := c.definitions.Get(def).(*definition)

	d := c.SetFactory(key, aliased.Factory)
	d.AliasOf = aliased

	return d
}

// AddProvider adds a new service provider.
func (c *containerBuilder) AddProvider(ps []Provider) {
	if len(ps) > 0 {
		c.panicIfResolved()
		c.providers = append(c.providers, ps...)
	}
}

// AddResolver adds a new service resolver.
func (c *containerBuilder) AddResolver(rs []Resolver) {
	if len(rs) > 0 {
		c.panicIfResolved()
		c.resolvers = append(c.resolvers, rs...)
	}
}

// GetContainer resolves and returns the corresponding container.
func (c *containerBuilder) GetContainer() *container {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.resolved {
		for _, p := range c.providers {
			p.Provide(c)
		}

		for _, r := range c.resolvers {
			r.Resolve(c)
		}

		c.resolved = true
	}

	return &container{
		builder:   c,
		instances: newItemHash(),
		sealed:    true,
		lock:      &sync.Mutex{},
	}
}

//GetTaggedKeys returns All keys related to a given tag. If values provided, then
//only the keys which match with tag and value will be returned.
func (c *containerBuilder) GetTaggedKeys(tag string, values []string) []string {
	tagged := make([]Binding, 0)
	for key, def := range c.definitions.All() {
		d := def.(*definition)
		if !d.Tags.Has(tag) {
			continue
		}
		if len(values) == 0 {
			tagged = append(tagged, Binding{key, d})
			continue
		}
		tagVal := d.Tags.Get(tag).(string)
		for _, v := range values {
			if v == tagVal {
				tagged = append(tagged, Binding{key, d})
				break
			}
		}
	}

	sort.SliceStable(tagged, func(i, j int) bool {
		return tagged[i].Target.(*definition).Priority > tagged[j].Target.(*definition).Priority
	})

	keys := make([]string, 0, len(tagged))
	for _, i := range tagged {
		keys = append(keys, i.Key)
	}
	return keys
}

func (c *containerBuilder) panicIfResolved() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.resolved {
		panic("container is resolved and new items can not be set")
	}
}
