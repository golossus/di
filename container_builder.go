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
	Register(builder *ContainerBuilder)
	Resolve(builder *ContainerBuilder)
}

type ContainerBuilder struct {
	params, defs, alias *itemHash
	providers           []Provider
	parser              *keyParser
	resolved            bool
}

func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		params:    newItemHash(),
		defs:      newItemHash(),
		alias:     newItemHash(),
		parser:    newKeyParser(),
		providers: make([]Provider, 0),
		resolved:  false,
	}
}

// Parameters:

func (c *ContainerBuilder) SetParameter(key string, param interface{}) {
	mustBeUnresolved(c)
	mustJsonMarshal(param)

	k, _ := c.parser.parse(key)

	c.params.set(k, param)
}

func (c *ContainerBuilder) HasParameter(key string) bool {
	return c.params.has(key)
}

func (c *ContainerBuilder) GetParameter(key string) interface{} {
	return c.params.get(key)
}

// Definitions:

func (c *ContainerBuilder) SetDefinition(key string, build interface{}) {
	mustBeUnresolved(c)
	mustBeValidConstructor(build)

	k, tags := c.parser.parse(key)

	c.alias.del(k)
	c.defs.set(k, &Definition{
		Build: build,
		Tags:  tags,
	})
}

func (c *ContainerBuilder) HasDefinition(key string) bool {
	if c.alias.has(key) {
		key = c.alias.get(key).(string)
	}

	return c.defs.has(key)
}

func (c *ContainerBuilder) GetDefinition(key string) *Definition {
	if c.alias.has(key) {
		key = c.alias.get(key).(string)
	}

	return c.defs.get(key).(*Definition)
}

// Aliases:

func (c *ContainerBuilder) SetAlias(key, def string) {
	mustBeUnresolved(c)

	if c.defs.has(key) {
		panic(fmt.Sprintf("definition with id '%s' already exists and alias cannot be set", key))
	}

	if !c.defs.has(def) {
		panic(fmt.Sprintf("definition with id '%s' does not exist and alias cannot be set", def))
	}

	c.alias.set(key, def)
}

func (c *ContainerBuilder) HasAlias(key string) bool {
	return c.alias.has(key)
}

func (c *ContainerBuilder) GetAlias(key string) string {
	return c.alias.get(key).(string)
}

// Providers

func (c *ContainerBuilder) AddProviders(p ...Provider) {
	c.providers = append(c.providers, p...)
}

func (c *ContainerBuilder) GetContainer() Container {
	//This should be protected about concurrency
	if c.resolved {
		return newContainer(c)
	}

	for _, p := range c.providers {
		p.Register(c)
	}

	for _, p := range c.providers {
		p.Resolve(c)
	}

	c.resolved = true

	return newContainer(c)
}

func mustBeUnresolved(c *ContainerBuilder) {
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

var cbType = reflect.TypeOf(&ContainerBuilder{})

func mustBeValidConstructor(build interface{}) {
	t := reflect.TypeOf(build)

	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("invalid constructor kind '%s'", t.Kind()))
	}

	if t.NumOut() != 1 {
		panic(fmt.Sprintf("constructor for type '%s' should return a single value", t.Name()))
	}
	if t.NumIn() != 1 || t.In(0) != cbType {
		panic(fmt.Sprintf("constructor for type '%s' should only receive a container builder instance", t.Name()))
	}
}
