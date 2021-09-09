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
	sealed bool
}

func (c *container) Get(key string) interface{} {
	def := c.builder.GetDefinition(key)
	if c.sealed && def.Private()  {
		panic(fmt.Sprintf("service with key '%s' is private and can't be retrieved from the container", key))
	}

	if def.Shared() && c.instances.has(key) {
		return c.instances.get(key)
	}

	c.instances.set(key, c.construct(def))

	return c.instances.get(key)
}

func (c *container) GetParameter(key string) interface{} {
	return c.builder.GetParameter(key)
}

func (c *container) construct(def *Definition) interface{} {
	args := []reflect.Value{}
	if reflect.TypeOf(def.Build).NumIn() > 0 {
		args = append(args, reflect.ValueOf(c.unseal()))
	}

	val := reflect.ValueOf(def.Build).Call(args)

	return val[0].Interface()
}

func (c *container) unseal() *container {
	if !c.sealed {
		return c
	}

	return  &container{
		builder: c.builder,
		instances: c.instances,
		sealed: false,
	}
}
