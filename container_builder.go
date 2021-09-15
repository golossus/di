// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"encoding/json"
	"fmt"
	"sync"
)

const (
	tagShared  = "shared"
	tagPrivate = "private"
)

//definition represents a service factory definition with additional metadata.
type definition struct {
	Factory         func(Container) interface{}
	Tags            *itemHash
	Shared, Private bool
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

//Resolver allows to resolve definitions into containerBuilder once All services
//definitions are available.
type Resolver interface {
	Resolve(builder ContainerBuilder)
}

//ContainerBuilder interface declares the public api for containerBuilder type.
type ContainerBuilder interface {
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

	c.definitions.set(k, &definition{
		Factory: factory,
		Tags:    tags,
		Shared:  tags.Has(tagShared),
		Private: tags.Has(tagPrivate),
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
	tagged := make([]string, 0)
	for key, def := range c.definitions.All() {
		d := def.(*definition)
		if !d.Tags.Has(tag) {
			continue
		}
		if len(values) == 0 {
			tagged = append(tagged, key)
			continue
		}
		tagVal := d.Tags.Get(tag).(string)
		for _, v := range values {
			if v == tagVal {
				tagged = append(tagged, key)
				break
			}
		}
	}

	return tagged
}

func (c *containerBuilder) panicIfResolved() {
	if c.resolved {
		panic("container is resolved and new items can not be set")
	}
}
