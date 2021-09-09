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
	shared := func(cb Container) int {
		spy++
		return spy
	}

	other := func(cb Container) int {
		return cb.Get("shared").(int)
	}

	b := NewContainerBuilder()
	b.SetDefinition("shared #shared", shared)
	b.SetDefinition("other", other)
	c := b.GetContainer()

	r1 := c.Get("shared").(int)
	r2 := c.Get("shared").(int)
	r3 := c.Get("other").(int)

	assert.Equal(t, 1, r1)
	assert.Equal(t, 1, r2)
	assert.Equal(t, 1, r3)
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

func TestContainer_Get_CanCreateWithPrivateDependency(t *testing.T) {
	public := func(cb Container) int {
		return cb.Get("public2").(int) + 1
	}

	public2 := func(cb Container) int {
		return cb.Get("private").(int) + 1
	}

	private := func() int {
		return 1
	}

	b := NewContainerBuilder()
	b.SetDefinition("public", public)
	b.SetDefinition("public2", public2)
	b.SetDefinition("private #private", private)
	c := b.GetContainer()

	r := c.Get("public").(int)

	assert.Equal(t, 3, r)
}

func TestContainer_Get_PanicsIfCircularReference(t *testing.T) {
	s1 := func(cb Container) int { return cb.Get("s2").(int) }
	s2 := func(cb Container) int { return cb.Get("s3").(int) }
	s3 := func(cb Container) int { return cb.Get("s1").(int) }

	b := NewContainerBuilder()
	b.SetDefinition("s1", s1)
	b.SetDefinition("s2", s2)
	b.SetDefinition("s3", s3)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "circular reference found while building service 's1' at service 's3'", func() {
		_ = c.Get("s1")
	})
}
