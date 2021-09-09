package di

import (
	"fmt"
	"reflect"
)

type container struct {
	builder   *containerBuilder
	instances *itemHash
}

func (c *container) Get(key string) interface{} {
	def := c.builder.GetDefinition(key)
	if def.Private() {
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
	if reflect.TypeOf(def.Build).NumIn() > 0{
		args = append(args, reflect.ValueOf(c))
	}

	val := reflect.ValueOf(def.Build).Call(args)

	return val[0].Interface()
}
