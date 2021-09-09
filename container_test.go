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

func TestContainer_Get_CanCreateWithContainerArgumentConstructor(t *testing.T) {
	newOne := func(cb Container) int {
		return 1
	}

	b := NewContainerBuilder()
	b.SetDefinition("one", newOne)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_CanCreateWithParameterDependencyConstructor(t *testing.T) {
	newOne := func(cb Container) int {
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
	newOne := func(cb Container) int {
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

func TestContainer_Get_CreatesSingleInstanceForSharedService(t *testing.T) {
	spy := 0
	newA := func(cb Container) int {
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

func TestContainer_Get_CreatesNewInstancesForNotSharedService(t *testing.T) {
	spy := 0
	newA := func(cb Container) int {
		spy++
		return spy
	}

	b := NewContainerBuilder()
	b.SetDefinition("a", newA)
	c := b.GetContainer()

	r1 := c.Get("a").(int)
	r2 := c.Get("a").(int)
	r3 := c.Get("a").(int)

	assert.Equal(t, 1, r1)
	assert.Equal(t, 2, r2)
	assert.Equal(t, 3, r3)
}

func TestContainer_Get_PanicsIfRequestingPrivateService(t *testing.T) {
	spy := 0
	newA := func(cb Container) int {
		spy++
		return spy
	}

	b := NewContainerBuilder()
	b.SetDefinition("a #private", newA)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "service with key 'a' is private and can't be retrieved from the container", func() {
		_ = c.Get("a").(int)
	})
}
