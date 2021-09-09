package di

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	tagShared  = "shared"
	tagPrivate = "private"
)

type Definition struct {
	Build interface{}
	Tags  *itemHash
}

func (d *Definition) Shared() bool {
	return d.Tags.has(tagShared)
}

func (d *Definition) Private() bool {
	return d.Tags.has(tagPrivate)
}

type Provider interface {
	Register(builder ContainerBuilder)
	Resolve(builder ContainerBuilder)
}

type Container interface {
	Get(key string) interface{}
	GetParameter(key string) interface{}
}

type ContainerBuilder interface {
	SetParameter(key string, param interface{})
	HasParameter(key string) bool
	GetParameter(key string) interface{}
	SetDefinition(key string, build interface{})
	HasDefinition(key string) bool
	GetDefinition(key string) *Definition
	SetAlias(key, def string)
	HasAlias(key string) bool
	GetAlias(key string) string
	AddProviders(p ...Provider)
}

type containerBuilder struct {
	params, defs, alias *itemHash
	providers           []Provider
	parser              *keyParser
	resolved            bool
}

func NewContainerBuilder() *containerBuilder {
	return &containerBuilder{
		params:    newItemHash(),
		defs:      newItemHash(),
		alias:     newItemHash(),
		parser:    newKeyParser(),
		providers: make([]Provider, 0),
		resolved:  false,
	}
}

// Parameters:

func (c *containerBuilder) SetParameter(key string, param interface{}) {
	mustBeUnresolved(c)
	mustJsonMarshal(param)

	k, _ := c.parser.parse(key)

	c.params.set(k, param)
}

func (c *containerBuilder) HasParameter(key string) bool {
	return c.params.has(key)
}

func (c *containerBuilder) GetParameter(key string) interface{} {
	return c.params.get(key)
}

// Definitions:

func (c *containerBuilder) SetDefinition(key string, build interface{}) {
	mustBeUnresolved(c)
	mustBeValidConstructor(build)

	k, tags := c.parser.parse(key)

	c.alias.del(k)
	c.defs.set(k, &Definition{
		Build: build,
		Tags:  tags,
	})
}

func (c *containerBuilder) HasDefinition(key string) bool {
	if c.alias.has(key) {
		key = c.alias.get(key).(string)
	}

	return c.defs.has(key)
}

func (c *containerBuilder) GetDefinition(key string) *Definition {
	if c.alias.has(key) {
		key = c.alias.get(key).(string)
	}

	return c.defs.get(key).(*Definition)
}

// Aliases:

func (c *containerBuilder) SetAlias(key, def string) {
	mustBeUnresolved(c)

	k, _ := c.parser.parse(key)

	if !c.defs.has(def) {
		panic(fmt.Sprintf("definition with id '%s' does not exist and alias cannot be set", def))
	}

	if c.defs.has(k) {
		panic(fmt.Sprintf("definition with id '%s' already exists and alias cannot be set", key))
	}

	c.alias.set(k, def)
}

func (c *containerBuilder) HasAlias(key string) bool {
	return c.alias.has(key)
}

func (c *containerBuilder) GetAlias(key string) string {
	return c.alias.get(key).(string)
}

// Providers

func (c *containerBuilder) AddProviders(p ...Provider) {
	if len(p) > 0 {
		mustBeUnresolved(c)
		c.providers = append(c.providers, p...)
	}
}

func (c *containerBuilder) GetContainer() *container {
	if !c.resolved {
		for _, p := range c.providers {
			p.Register(c)
		}

		for _, p := range c.providers {
			p.Resolve(c)
		}

		c.resolved = true
	}

	return &container{
		builder:   c,
		instances: newItemHash(),
	}
}

func mustBeUnresolved(c *containerBuilder) {
	if c.resolved {
		panic("container is resolved and new items can not be set")
	}
}

func mustJsonMarshal(param interface{}) {
	_, err := json.Marshal(param)
	if err != nil {
		panic(fmt.Sprintf("invalid parameter param '%#v'", param))
	}
}

var cbType = reflect.TypeOf((*Container)(nil)).Elem()

func mustBeValidConstructor(build interface{}) {
	t := reflect.TypeOf(build)

	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("invalid constructor kind '%T', must be a function", build))
	}

	if t.NumOut() != 1 {
		panic(fmt.Sprintf("constructor '%T' should return a single value", build))
	}

	if t.NumIn() == 0 {
		return
	}

	if t.NumIn() > 1 || !t.In(0).Implements(cbType) {
		panic(fmt.Sprintf("constructor '%T' can only receive a '%s' argument", build, cbType.Name()))
	}
}
