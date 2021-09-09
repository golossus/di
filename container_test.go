package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainer_Get_CanCreateWithZeroArgumentsConstructor(t *testing.T) {
	newOne := func() int {
		return 1
	}

	b := NewContainerBuilder()
	b.SetDefinition("one", newOne)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_CanCreateWithBuilderArgumentConstructor(t *testing.T) {
	newOne := func(cb ContainerInterface) int {
		return 1
	}

	b := NewContainerBuilder()
	b.SetDefinition("one", newOne)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_CanCreateWithParameterDependencyConstructor(t *testing.T) {
	newOne := func(cb ContainerInterface) int {
		return cb.GetParameter("one").(int)
	}

	b := NewContainerBuilder()
	b.SetDefinition("one", newOne)
	b.SetParameter("one", 1)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_CanCreateWithDefinitionDependencyConstructor(t *testing.T) {
	newOne := func(cb ContainerInterface) int {
		return cb.Get("two").(int) - 1
	}

	newTwo := func() int {
		return 2
	}

	b := NewContainerBuilder()
	b.SetDefinition("one", newOne)
	b.SetDefinition("two", newTwo)
	b.SetParameter("one", 1)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_SharedService(t *testing.T) {
	spy := 0
	newA := func(cb ContainerInterface) int {
		spy++
		return spy
	}

	b := NewContainerBuilder()
	b.SetDefinition("a #shared", newA)
	c := b.GetContainer()

	r1 := c.Get("a").(int)
	r2 := c.Get("a").(int)

	assert.Equal(t, r1, r2)
}
