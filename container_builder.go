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
	TagShared   = "shared"
	TagPrivate  = "private"
	TagPriority = "priority"
	TagInject   = "inject"
	TagValue    = "value"
	TagAlias    = "alias"
	TagFactory  = "factory"
)

type Binding struct {
	Key    string
	Target interface{}
	Tags   map[string]string
}

// Provider allows providing definitions into containerBuilder. Binding dependencies
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

// ContainerBuilder interface declares the public API for containerBuilder type. ContainerBuilder is used
// to declare all the services such as values, factories, injectables or aliases; to build an IOC or dependency
// injection container.
//
// The setter methods allow providing extra tags to add metadata to service definitions. There are some reserved
// tags which are useful to indicate the container how to build a specific service:
//
//	- TagShared: default "true", declares a service as "shared" and the container will return a singleton. It is
//	  required to use pointers for returned services to work as real singleton.
// 	- TagPrivate: default "true", declares a service as "private" and it will be available to be injected as dependency
//	  of another service but not available to be retrieved from current container.
//	- TagValue, TagFactory, TagInject and TagAlias: default "", these tags are mainly used with bindings and SetAll
//	  method to indicate the container which kind of service to use (a value, a factory, an injectable struct or an alias).
//	- TagPriority: default "0", can be used to sort the services by priority when retrieving services by tag. The higher
//	  the value, the higher the priority. Services will be sorted and the ones with higher priority will be returned
//	  on the lowest indexes of the result slice.
//
// By default, services can be overwritten by using the same key as an existing one. Aliases can also be overwritten but
// trying to set an alias with a key used by a real service definition will fail.
type ContainerBuilder interface {
	SetAll(all []Binding)
	SetValue(key string, value interface{}, tags ...map[string]string) *definition
	SetFactory(key string, factory func(Container) interface{}, tags ...map[string]string) *definition
	SetInjectable(key string, value interface{}, tags ...map[string]string) *definition
	SetAlias(key, def string, tags ...map[string]string) *definition
	HasDefinition(key string) bool
	GetDefinition(key string) *definition
	GetTaggedKeys(tag string, values []string) []string
	GetContainer() *container
}

// containerBuilder implements ContainerBuilder interface to bind service definitions
// and resolve the final service container.
type containerBuilder struct {
	definitions *itemHash
	providers   []Provider
	resolvers   []Resolver
	resolved    bool
	lock        *sync.Mutex
}

// NewContainerBuilder returns a pointer to a new containerBuilder instance.
func NewContainerBuilder() *containerBuilder {
	return &containerBuilder{
		definitions: newItemHash(),
		providers:   make([]Provider, 0),
		resolvers:   make([]Resolver, 0),
		resolved:    false,
		lock:        &sync.Mutex{},
	}
}

func (c *containerBuilder) panicIfResolved() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.resolved {
		panic("container is resolved and new items can not be set")
	}
}

func (c *containerBuilder) setDefinition(key string, factory func(c Container) interface{}, tags ...map[string]string) *definition {
	c.panicIfResolved()

	k, t := parseKey(key)

	tags = append(tags, t)
	def, err := newDefinition(factory, tags...)
	if err != nil {
		panic(fmt.Sprintf("%s for key '%s'", err, k))
	}
	c.definitions.set(k, def)

	return def
}

// SetValue adds a new value or instance definition to the container on a given Key. When retrieving from the container
// by the given key, it will always return the given value.
func (c *containerBuilder) SetValue(key string, value interface{}, tags ...map[string]string) *definition {
	tags = append(tags, map[string]string{TagValue: ""})
	return c.setDefinition(key, func(_ Container) interface{} {
		return value
	}, tags...)
}

// SetFactory adds a new factory definition to the container referenced by a given Key. When retrieving from the container
// by the given key, the container will call this factory to create the corresponding service.
func (c *containerBuilder) SetFactory(key string, factory func(Container) interface{}, tags ...map[string]string) *definition {
	tags = append(tags, map[string]string{TagFactory: ""})
	return c.setDefinition(key, factory, tags...)
}

// SetInjectable adds a new injectable struct definition to the container on a given key. Given struct must contain at
// least one public member labeled with the "inject" label:
//
//	 type SomeType struct {
//	 	 field string `inject:"service.to.inject.key"`
//	 }
//
// As shown in the example above, with the "inject" label we can configure the dependencies of the injectable service by
// indicating the key of the required dependency. When retrieving this service by the given key, the container will
// inject the indicated dependencies.
func (c *containerBuilder) SetInjectable(key string, i interface{}, tags ...map[string]string) *definition {
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

	tags = append(tags, map[string]string{TagInject: ""})
	d := c.setDefinition(key, func(c Container) interface{} {
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
	}, tags...)

	return d

}

// SetAlias sets an alias for an existing definition on a given key. Aliases inherit the aliased service factory, but
// they can have their own set of tags. As an example, a service might be "private" and the corresponding alias can be
// public or even a singleton. Aliases can be replaced by real services definitions, the contrary will fail.
func (c *containerBuilder) SetAlias(key, def string, tags ...map[string]string) *definition {

	if !c.definitions.Has(def) {
		panic(fmt.Sprintf("definition with id '%s' does not exist and alias cannot be set", def))
	}

	if c.definitions.Has(key) && c.definitions.Get(key).(*definition).AliasOf == nil {
		panic(fmt.Sprintf("definition with id '%s' already exists and alias cannot be set", key))
	}

	aliased := c.definitions.Get(def).(*definition)

	tags = append(tags, map[string]string{TagAlias: ""})
	d := c.setDefinition(key, aliased.Factory, tags...)
	d.AliasOf = aliased

	return d
}

// SetAll adds given bindings into the containerBuilder. Reserved tags TagValue, TagAlias, TagFactory and TagInject;
// are used to determine the kind of service definition to consider for each Binding. By default, TagFactory is used
// if no other kind is indicated. Commented tags are all mutually exclusive and adding more than one per Binding will
// fail.
//
//	b.SetAll([]Binding{
//		{Key: "key1 #factory", Target: func(c Container) interface{} {
//			return 1
//		}, Tags: map[string]string{TagFactory: ""}},
//		{Key: "key2", Target: 2, Tags: map[string]string{TagValue: ""}},
//		{Key: "key3", Target: "key2", Tags: map[string]string{TagAlias: ""}},
//		{Key: "key4", Target: struct{}{}, Tags: map[string]string{TagInject: ""}},
//		{Key: "key5", Target: func(c Container) interface{} {  	// <- defaults to TagFactory
//			return 5
//		}},
//		{Key: "key4", Target: "key2, Tags: map[string]string{TagValue: "", TagFactory: ""}}, 	// <- will panic
//	})
func (c *containerBuilder) SetAll(all []Binding) {
	for _, b := range all {
		k, parsedTags := parseKey(b.Key)
		mergedTags := mergeTags(b.Tags, parsedTags)

		kind, err := selectKindTag(mergedTags)
		if err != nil {
			panic(fmt.Sprintf("%s for key '%s'", err, k))
		}

		switch kind{
		case TagAlias:
			c.SetAlias(k, b.Target.(string), mergedTags)
		case TagValue:
			c.SetValue(k, b.Target, mergedTags)
		case TagInject:
			c.SetInjectable(k, b.Target, mergedTags)
		case TagFactory:
			fallthrough
		default:
			c.SetFactory(k, b.Target.(func(Container) interface{}), mergedTags)
		}
	}
}

// HasDefinition returns true if definition for the Key exists in the container.
func (c *containerBuilder) HasDefinition(key string) bool {
	return c.definitions.Has(key)
}

// GetDefinition retrieves a container definition for the Key or panics if not found.
func (c *containerBuilder) GetDefinition(key string) *definition {
	return c.definitions.Get(key).(*definition)
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
		tagVal, ok := d.Tags[tag]
		if !ok {
			continue
		}

		if len(values) == 0 {
			tagged = append(tagged, Binding{Key: key, Target: d})
			continue
		}

		for _, v := range values {
			if v == tagVal {
				tagged = append(tagged, Binding{Key: key, Target: d})
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
