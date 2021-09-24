// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	tagShared   = "shared"
	tagPrivate  = "private"
	tagPriority = "priority"
	tagInject   = "inject"

	paramPrefix = "_"

	priorityDefault = 0
)

type Some struct {
	Key string
	Val interface{}
}

//definition represents a service factory definition with additional metadata.
type definition struct {
	Factory         func(Container) interface{}
	Tags            *itemHash
	Shared, Private bool
	Priority        int16
}

//alias represents a service alias identifier and additional metadata.
type alias struct {
	Key     string
	Private bool
}

//Provider allows to provide definitions into containerBuilder. Some dependencies
//might not be available during the call to this method.
type Provider interface {
	Provide(builder ContainerBuilder)
}

//ProviderFunc adapts a normal func into a Provider.
type ProviderFunc func(ContainerBuilder)

func (f ProviderFunc) Provide(b ContainerBuilder) {
	f(b)
}

//Resolver allows to resolve definitions into containerBuilder once All services
//definitions are available.
type Resolver interface {
	Resolve(builder ContainerBuilder)
}

//ProviderFunc adapts a normal func into a Resolver.
type ResolverFunc func(ContainerBuilder)

func (f ResolverFunc) Resolve(b ContainerBuilder) {
	f(b)
}

//ContainerBuilder interface declares the public api for containerBuilder type.
type ContainerBuilder interface {
	SetMany(all ...Some)
	SetDefinition(key string, factory func(c Container) interface{})
	HasDefinition(key string) bool
	GetDefinition(key string) *definition
	SetParameter(key string, value interface{})
	HasParameter(key string) bool
	GetParameter(key string) interface{}
	SetAlias(key, def string)
	HasAlias(key string) bool
	GetAlias(key string) *alias
	GetTaggedKeys(tag string, values []string) []string
}

type containerBuilder struct {
	definitions, parameters, alias *itemHash
	parser                         *keyParser
	providers                      []Provider
	resolvers                      []Resolver
	resolved                       bool
	lock                           *sync.Mutex
}

//NewContainerBuilder returns a pointer to a new containerBuilder instance.
func NewContainerBuilder() *containerBuilder {
	return &containerBuilder{
		parameters:  newItemHash(),
		definitions: newItemHash(),
		alias:       newItemHash(),
		parser:      newKeyParser(),
		providers:   make([]Provider, 0),
		resolvers:   make([]Resolver, 0),
		resolved:    false,
		lock:        &sync.Mutex{},
	}
}

//SetParameter adds a new parameter value to the container on a given Key.
func (c *containerBuilder) SetParameter(key string, value interface{}) {
	c.panicIfResolved()

	if _, err := json.Marshal(value); err != nil {
		panic(fmt.Sprintf("invalid parameter param '%#v'", value))
	}

	k, _ := c.parser.parse(key)
	c.parameters.set(k, value)
}

//HasParameter returns true if parameter for the Key exists in the container.
func (c *containerBuilder) HasParameter(key string) bool {
	return c.parameters.Has(key)
}

//GetParameter retrieves a container parameter for the Key or panics if not found.
func (c *containerBuilder) GetParameter(key string) interface{} {
	return c.parameters.Get(key)
}

func (c *containerBuilder) SetMany(all ...Some) {
	for _, i := range all {
		switch i.Val.(type) {
		case string:
			c.SetAlias(i.Key, i.Val.(string))
		case func(c Container) interface{}:
			c.SetDefinition(i.Key, i.Val.(func(c Container) interface{}))
		default:
			_, tags := c.parser.parse(i.Key)
			if tags.Has(tagInject) {
				c.SetInjectable(i.Key, i.Val)
				continue
			}
			c.SetParameter(i.Key, i.Val)
		}
	}
}

func (c *containerBuilder) SetInjectable(key string, i interface{}) {
	t := reflect.TypeOf(i)
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		t = t.Elem()
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

	c.SetDefinition(key, func(c Container) interface{} {
		t := reflect.New(t)
		e := t.Elem()
		for i, k := range fields {
			var p interface{}
			if strings.HasPrefix(k, paramPrefix) {
				p = c.GetParameter(strings.TrimLeft(k, paramPrefix))
			} else {
				p = c.Get(k)
			}

			v := reflect.ValueOf(p)
			e.Field(i).Set(v)
		}

		if isPtr {
			return t.Interface()

		}

		return e.Interface()
	})

}

//SetDefinition adds a new definition to the container referenced by a given
//Key. Keys can contain tags in the form of "#tag[=value]" where the value part
//can be omitted. Reserved tags "#shared" and "#private" can be used to make a
//definition shared (factory will return a singleton) and private (service will
//be available to be injected as dependency but not available to be retrieved
//from current container).
func (c *containerBuilder) SetDefinition(key string, factory func(c Container) interface{}) {
	c.panicIfResolved()

	k, tags := c.parser.parse(key)

	c.alias.del(k)

	priority := int16(priorityDefault)
	if tags.Has(tagPriority) {
		prioValue := tags.Get(tagPriority).(string)
		parsed, err := strconv.ParseInt(prioValue, 10, 16)
		if err != nil {
			panic(fmt.Sprintf("priority value %s is not a valid number", prioValue))
		}
		priority = int16(parsed)
	}

	c.definitions.set(k, &definition{
		Factory:  factory,
		Tags:     tags,
		Shared:   tags.Has(tagShared),
		Private:  tags.Has(tagPrivate),
		Priority: priority,
	})
}

//HasDefinition returns true if definition for the Key exists in the container.
func (c *containerBuilder) HasDefinition(key string) bool {
	if c.alias.Has(key) {
		key = c.alias.Get(key).(*alias).Key
	}

	return c.definitions.Has(key)
}

//GetDefinition retrieves a container definition for the Key or panics if not found.
func (c *containerBuilder) GetDefinition(key string) *definition {
	if !c.alias.Has(key) {
		return c.definitions.Get(key).(*definition)
	}
	a := c.alias.Get(key).(*alias)
	d := c.definitions.Get(a.Key).(*definition)

	if d.Private == a.Private {
		return d
	}

	return &definition{
		Factory: d.Factory,
		Tags:    d.Tags,
		Private: a.Private,
	}
}

//SetAlias sets an alias for an existing definition.
func (c *containerBuilder) SetAlias(key, def string) {
	c.panicIfResolved()

	k, tags := c.parser.parse(key)

	if !c.definitions.Has(def) {
		panic(fmt.Sprintf("definition with id '%s' does not exist and alias cannot be set", def))
	}

	if c.definitions.Has(k) {
		panic(fmt.Sprintf("definition with id '%s' already exists and alias cannot be set", key))
	}

	c.alias.set(k, &alias{def, tags.Has(tagPrivate)})
}

//HasAlias returns true if given alias Has been set into the container.
func (c *containerBuilder) HasAlias(key string) bool {
	return c.alias.Has(key)
}

//GetAlias returns the service Key related to given alias Key.
func (c *containerBuilder) GetAlias(key string) *alias {
	return c.alias.Get(key).(*alias)
}

//AddProvider adds a new service provider.
func (c *containerBuilder) AddProvider(ps []Provider) {
	if len(ps) > 0 {
		c.panicIfResolved()
		c.providers = append(c.providers, ps...)
	}
}

//AddResolver adds a new service resolver.
func (c *containerBuilder) AddResolver(rs []Resolver) {
	if len(rs) > 0 {
		c.panicIfResolved()
		c.resolvers = append(c.resolvers, rs...)
	}
}

//GetContainer resolves and returns the corresponding container.
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
	tagged := make([]Some, 0)
	for key, def := range c.definitions.All() {
		d := def.(*definition)
		if !d.Tags.Has(tag) {
			continue
		}
		if len(values) == 0 {
			tagged = append(tagged, Some{key, d})
			continue
		}
		tagVal := d.Tags.Get(tag).(string)
		for _, v := range values {
			if v == tagVal {
				tagged = append(tagged, Some{key, d})
				break
			}
		}
	}

	sort.SliceStable(tagged, func(i, j int) bool {
		return tagged[i].Val.(*definition).Priority > tagged[j].Val.(*definition).Priority
	})

	keys := make([]string, 0, len(tagged))
	for _, i := range tagged{
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
