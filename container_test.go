// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainer_Get_BuildsService(t *testing.T) {
	f := func(cb Container) interface{} { return 1 }
	k := "one"

	b := NewContainerBuilder()
	b.SetFactory(k, f)
	c := b.GetContainer()

	r := c.Get(k).(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_BuildsParameterDependencies(t *testing.T) {
	f := func(cb Container) interface{} {
		return cb.Get("one").(int)
	}

	b := NewContainerBuilder()
	b.SetFactory("one", f)
	b.SetParameter("one", 1)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_BuildsDefinitionDependencies(t *testing.T) {
	f := func(cb Container) interface{} {
		return cb.Get("two").(int) - 1
	}

	newTwo := func(_ Container) interface{} {
		return 2
	}

	b := NewContainerBuilder()
	b.SetFactory("one", f)
	b.SetFactory("two", newTwo)
	b.SetParameter("one", 1)
	c := b.GetContainer()

	r := c.Get("one").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_CreatesSingleInstanceForSharedService(t *testing.T) {
	spy := 0
	shared := func(cb Container) interface{} {
		spy++
		return spy
	}

	other := func(cb Container) interface{} {
		return cb.Get("shared").(int)
	}

	b := NewContainerBuilder()
	b.SetFactory("shared #shared", shared)
	b.SetFactory("other", other)
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
	newA := func(cb Container) interface{} {
		spy++
		return spy
	}

	b := NewContainerBuilder()
	b.SetFactory("a", newA)
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
	newA := func(cb Container) interface{} {
		spy++
		return spy
	}

	b := NewContainerBuilder()
	b.SetFactory("a #private", newA)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "service with key 'a' is private and can't be retrieved from the container", func() {
		_ = c.Get("a").(int)
	})
}

func TestContainer_Get_PanicsIfRequestingPrivateAlias(t *testing.T) {
	spy := 0
	newA := func(cb Container) interface{} { spy++; return spy }

	b := NewContainerBuilder()
	b.SetFactory("a", newA)
	b.SetAlias("a_alias #private", "a")
	c := b.GetContainer()

	assert.PanicsWithValue(t, "service with key 'a_alias' is private and can't be retrieved from the container", func() {
		_ = c.Get("a_alias").(int)
	})
}

func TestContainer_Get_RetrievesPrivateServiceThroughPublicAlias(t *testing.T) {
	spy := 0
	newA := func(cb Container) interface{} { spy++; return spy }

	b := NewContainerBuilder()
	b.SetFactory("a #private ", newA)
	b.SetAlias("a_alias", "a")
	c := b.GetContainer()

	s := c.Get("a_alias").(int)

	assert.Equal(t, 1, s)
}

func TestContainer_Get_CanCreateWithPrivateDependency(t *testing.T) {
	public := func(cb Container) interface{} {
		return cb.Get("public2").(int) + 1
	}

	public2 := func(cb Container) interface{} {
		return cb.Get("private").(int) + 1
	}

	private := func(_ Container) interface{} {
		return 1
	}

	b := NewContainerBuilder()
	b.SetFactory("public", public)
	b.SetFactory("public2", public2)
	b.SetFactory("private #private", private)
	c := b.GetContainer()

	r := c.Get("public").(int)

	assert.Equal(t, 3, r)
}

func TestContainer_Get_CanCreateWithPrivateAliasDependency(t *testing.T) {
	public := func(cb Container) interface{} {
		return cb.Get("public2").(int) + 1
	}

	public2 := func(cb Container) interface{} {
		return cb.Get("private_alias").(int) + 1
	}

	private := func(_ Container) interface{} {
		return 1
	}

	b := NewContainerBuilder()
	b.SetFactory("public", public)
	b.SetFactory("public2", public2)
	b.SetFactory("private #private", private)
	b.SetAlias("private_alias #private", "private")
	c := b.GetContainer()

	r := c.Get("public").(int)

	assert.Equal(t, 3, r)
}

func TestContainer_Get_CanCreateWithPrivateTaggedDependencies(t *testing.T) {
	plus1 := func(cb Container) interface{} {
		return cb.Get("sum").(int) + 1
	}

	sum := func(cb Container) interface{} {
		sum := 0
		for _, s := range cb.GetTaggedBy("sum") {
			sum += s.(int)
		}
		return sum
	}

	tagged1 := func(_ Container) interface{} {
		return 1
	}
	tagged2 := func(_ Container) interface{} {
		return 10
	}
	tagged3 := func(_ Container) interface{} {
		return 100
	}

	b := NewContainerBuilder()
	b.SetFactory("plus1", plus1)
	b.SetFactory("sum #private", sum)
	b.SetFactory("tagged1 #sum", tagged1)
	b.SetFactory("tagged2 #sum #shared #private", tagged2)
	b.SetFactory("tagged3 #sum #private", tagged3)
	c := b.GetContainer()

	r := c.Get("plus1").(int)

	assert.Equal(t, 112, r)
}

func TestContainer_Get_PanicsIfCircularReference(t *testing.T) {
	s1 := func(cb Container) interface{} { return cb.Get("s2").(int) }
	s2 := func(cb Container) interface{} { return cb.Get("s3").(int) }
	s3 := func(cb Container) interface{} { return cb.Get("s1").(int) }

	b := NewContainerBuilder()
	b.SetFactory("s1", s1)
	b.SetFactory("s2", s2)
	b.SetFactory("s3", s3)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "circular reference found while building service 's1' at service 's3'", func() {
		_ = c.Get("s1")
	})
}
func TestContainer_GetTaggedBy(t *testing.T) {

	tagged1 := func(_ Container) interface{} { return 1 }
	tagged2 := func(_ Container) interface{} { return 10 }
	tagged3 := func(_ Container) interface{} { return 100 }

	t.Run("retrieves services ordered by priority", func(t *testing.T) {
		b := NewContainerBuilder()
		b.SetFactory("tagged1 #sum #priority=1", tagged1)
		b.SetFactory("tagged2 #sum", tagged2)
		b.SetFactory("tagged3 #sum #priority=2", tagged3)

		c := b.GetContainer()

		result := c.GetTaggedBy("sum")

		expected := []interface{}{100, 1, 10}
		assert.Equal(t, expected, result)
	})

}

func TestContainer_GetTaggedBy_PanicsIfSomeServiceIsPrivate(t *testing.T) {

	tagged1 := func(_ Container) interface{} {
		return 1
	}
	tagged2 := func(_ Container) interface{} {
		return 10
	}
	tagged3 := func(_ Container) interface{} {
		return 100
	}

	b := NewContainerBuilder()
	b.SetFactory("tagged1 #sum", tagged1)
	b.SetFactory("tagged2 #sum #shared", tagged2)
	b.SetFactory("tagged3 #sum #private", tagged3)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "service with key 'tagged3' is private and can't be retrieved from the container", func() {
		_ = c.GetTaggedBy("sum")
	})
}

func TestContainer_GetTaggedBy_PanicsIfCircularReferenceUsedForDependencies(t *testing.T) {
	s1 := func(cb Container) interface{} { return cb.GetTaggedBy("tag")[0].(int) }
	s2 := func(cb Container) interface{} { return cb.Get("s3").(int) }
	s3 := func(cb Container) interface{} { return cb.Get("s1").(int) }

	b := NewContainerBuilder()
	b.SetFactory("s1", s1)
	b.SetFactory("s2 #tag", s2)
	b.SetFactory("s3 #tag", s3)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "circular reference found while building service 's1' at service 's3'", func() {
		_ = c.Get("s1")
	})
}

func TestContainer_GetTaggedBy_PanicsIfCircularReference(t *testing.T) {
	s1 := func(cb Container) interface{} { return cb.Get("s2").(int) }
	s2 := func(cb Container) interface{} { return cb.Get("s3").(int) }
	s3 := func(cb Container) interface{} { return cb.Get("s1").(int) }

	b := NewContainerBuilder()
	b.SetFactory("s1", s1)
	b.SetFactory("s2 #tag=2", s2)
	b.SetFactory("s3 #tag=3", s3)
	c := b.GetContainer()

	assert.PanicsWithValue(t, "circular reference found while building service 's3' at service 's2'", func() {
		_ = c.GetTaggedBy("tag", "3")
	})
	assert.PanicsWithValue(t, "circular reference found while building service 's2' at service 's1'", func() {
		_ = c.GetTaggedBy("tag", "2")
	})
}

func TestContainer_Get_SharedServiceIfConcurrentAccess(t *testing.T) {
	init := 0
	seed := &init
	s1 := func(cb Container) interface{} {
		*seed = *seed + 1
		return seed
	}

	b := NewContainerBuilder()
	b.SetFactory("s1 #shared", s1)
	c := b.GetContainer()

	for i := 0; i < 1000; i++ {
		go func() {
			actual := c.Get("s1").(*int)
			assert.Equal(t, 1, *actual)
		}()
	}
}

func TestContainer_MustBuild_panicsOnInvalidService(t *testing.T) {
	s1 := func(c Container) interface{} {
		return c.Get("s2").(int)
	}
	s2 := func(c Container) interface{} {
		return "I'm a string!'"
	}

	b := NewContainerBuilder()
	b.SetFactory("s1", s1)
	b.SetFactory("s2", s2)

	c := b.GetContainer()

	assert.Panics(t, func() {
		c.MustBuild(true)
	})
}

func TestContainer_MustBuild_doesNotPanicAndClearsInstancesIfAllValid(t *testing.T) {
	s1 := func(c Container) interface{} {
		return c.Get("s2").(*int)
	}

	s2 := func(c Container) interface{} {
		value := 1
		return &value
	}

	b := NewContainerBuilder()
	b.SetFactory("s1", s1)
	b.SetFactory("s2", s2)
	c := b.GetContainer()

	c.MustBuild(true)

	assert.Empty(t, c.instances.All())
}

func TestContainer_MustBuild_doesNotPanicAndDoesNotClearInstancesIfAllValid(t *testing.T) {
	s1 := func(c Container) interface{} {
		return c.Get("s2").(*int)
	}

	s2 := func(c Container) interface{} {
		value := 1
		return &value
	}

	b := NewContainerBuilder()
	b.SetFactory("s1", s1)
	b.SetFactory("s2 #shared #private", s2)
	c := b.GetContainer()

	c.MustBuild(false)

	assert.Len(t, c.instances.All(), 1)
}
