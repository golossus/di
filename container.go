package di

import (
	"fmt"
	"reflect"
)

type Container struct {
	builder   *ContainerBuilder
	instances *itemHash
}

func newContainer(builder *ContainerBuilder) Container {
	return Container{
		builder:   builder,
		instances: newItemHash(),
	}
}

func (c Container) Has(key string) bool {
	return c.builder.HasDefinition(key)
}

func (c Container) Get(key string) interface{} {
	return c.create(key)
}

func (c Container) HasParameter(key string) bool {
	return c.builder.HasParameter(key)
}

func (c Container) GetParameter(key string) interface{} {
	return c.builder.GetParameter(key)
}

func (c Container) create(key string) interface{} {
	def := c.builder.GetDefinition(key)
	if def.Private() {
		panic(fmt.Sprintf("service with key '%s' is private and can't be retrieved from the container", key))
	}

	if def.Shared() && c.instances.has(key) {
		return c.instances.get(key)
	}

	args := []reflect.Value{reflect.ValueOf(c.builder)}
	val := reflect.ValueOf(def.Build).Call(args)

	c.instances.set(key, val[0].Interface())

	return c.instances.get(key)
}
