package di

import (
	"fmt"
	"reflect"
)

type Container interface {
	Get(key string) interface{}
	GetParameter(key string) interface{}
}

type container struct {
	builder   *containerBuilder
	instances *itemHash
	sealed    bool
	loading   []string
}

func (c *container) Get(key string) interface{} {
	def := c.builder.GetDefinition(key)
	if c.sealed && def.Private() {
		panic(fmt.Sprintf("service with key '%s' is private and can't be retrieved from the container", key))
	}

	if def.Shared() && c.instances.has(key) {
		return c.instances.get(key)
	}

	c.instances.set(key, c.construct(def, key))

	return c.instances.get(key)
}

func (c *container) GetParameter(key string) interface{} {
	return c.builder.GetParameter(key)
}

func (c *container) construct(def *Definition, key string) interface{} {
	for i := 0; i < len(c.loading); i++ {
		if c.loading[i] == key {
			msg := "circular reference found while building service '%s' at service '%s'"
			panic(fmt.Sprintf(msg, c.loading[0], c.loading[len(c.loading)-1]))
		}
	}

	args := []reflect.Value{}
	if reflect.TypeOf(def.Build).NumIn() > 0 {
		u := c.unseal()
		u.loading = append(u.loading, key)
		args = append(args, reflect.ValueOf(u))
	}

	val := reflect.ValueOf(def.Build).Call(args)

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
	}
}
