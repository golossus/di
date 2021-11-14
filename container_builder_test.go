// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewContainerBuilder_ReturnsInitialised(t *testing.T) {
	b := NewContainerBuilder()
	assert.NotNil(t, b.parameters)
	assert.NotNil(t, b.definitions)
	assert.NotNil(t, b.providers)
	assert.NotNil(t, b.parser)
	assert.Len(t, b.parameters.All(), 0)
	assert.Len(t, b.definitions.All(), 0)
	assert.Len(t, b.providers, 0)
	assert.False(t, b.resolved)
}

func TestContainerBuilder_SetParameter_IfValidValue(t *testing.T) {
	data := map[string]interface{}{
		"booleans": true,
		"strings":  "abc",
		"integers": 123,
		"floats":   1.23,
		"slices":   []byte{'1', '2', '3'},
		"maps":     map[string]int16{"1": 1, "2": 2, "3": 3},
		"struct": struct {
			Greet  string
			Number int
		}{
			"hello",
			123,
		},
	}

	for key, val := range data {
		t.Run(key, func(t *testing.T) {
			b := NewContainerBuilder()
			b.SetParameter(key, val)

			assert.True(t, b.HasParameter(key))
			assert.Equal(t, val, b.GetParameter(key))
		})
	}
}

func TestContainerBuilder_SetParameter_IgnoresKeyTags(t *testing.T) {
	key := "some-Key-with #tag"
	keyParsed := "some-Key-with"
	val := 1

	b := NewContainerBuilder()
	b.SetParameter(key, val)

	assert.False(t, b.HasParameter(key))
	assert.True(t, b.HasParameter(keyParsed))
	assert.Equal(t, val, b.GetParameter(keyParsed))
}

func TestContainerBuilder_SetParameter_PanicsIfInvalidParameterValue(t *testing.T) {
	b := NewContainerBuilder()
	param := make(chan int)

	assert.PanicsWithValue(t, fmt.Sprintf("invalid parameter param '%#v'", param), func() {
		b.SetParameter("id", param)
	})
}

func TestContainerBuilder_SetParameter_PanicsIfResolved(t *testing.T) {
	b := NewContainerBuilder()
	b.resolved = true

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.SetParameter("id", 1)
	})
}

func TestContainerBuilder_SetDefinition_PanicsIfResolved(t *testing.T) {
	b := NewContainerBuilder()
	b.resolved = true

	key := "key1"
	val := func(c Container) interface{} {
		return 1
	}

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.SetDefinition(key, val)
	})
}

func TestContainerBuilder_SetDefinition_Success(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	assert.True(t, b.HasDefinition(key))

	f := b.GetDefinition(key).Factory
	assert.Equal(t, val(&container{}), f(&container{}))
}

func TestContainerBuilder_SetDefinition_SuccessIfConstructorHasNoArgument(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	val := func(_ Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	assert.True(t, b.HasDefinition(key))

	f := b.GetDefinition(key).Factory
	assert.Equal(t, val(&container{}), f(&container{}))
}

func TestContainerBuilder_SetDefinition_RemovesAlias(t *testing.T) {
	b := NewContainerBuilder()
	alias := "alias"
	key := "Key"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasDefinition(alias))

	k := b.GetDefinition(key)
	a := b.GetDefinition(alias)
	assert.Equal(t, k.Factory(&container{}), a.Factory(&container{}))
	assert.Equal(t, k, a.AliasOf)

	b.SetDefinition(alias, val)
	assert.True(t, b.HasDefinition(alias))

	a = b.GetDefinition(alias)
	assert.Equal(t, val(&container{}), a.Factory(&container{}))
	assert.Nil(t, a.AliasOf)
}

func TestContainerBuilder_SetDefinition_SuccessWithTags(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1 #private #shared #other"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	assert.False(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition("key1"))

	d := b.GetDefinition("key1")
	assert.True(t, d.Private)
	assert.True(t, d.Shared)
	assert.True(t, d.Tags.Has("other"))

	f := d.Factory
	assert.Equal(t, val(&container{}), f(&container{}))
}

func TestContainerBuilder_SetAlias_PanicsIfResolved(t *testing.T) {
	b := NewContainerBuilder()
	b.resolved = true

	key := "key1"
	key2 := "key2"

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.SetAlias(key, key2)
	})
}

func TestContainerBuilder_SetAlias_PanicsIfServiceDoesNotExist(t *testing.T) {
	b := NewContainerBuilder()

	key := "key1"
	def := "def1"

	assert.PanicsWithValue(t, "definition with id 'def1' does not exist and alias cannot be set", func() {
		b.SetAlias(key, def)
	})
}

func TestContainerBuilder_SetAlias_PanicsIfServiceKeyAlreadyExists(t *testing.T) {
	b := NewContainerBuilder()

	key := "key1"
	def := "key1"
	val := func(c Container) interface{} {
		return 1
	}
	b.SetDefinition(def, val)

	assert.PanicsWithValue(t, "definition with id 'key1' already exists and alias cannot be set", func() {
		b.SetAlias(key, def)
	})
}

func TestContainerBuilder_SetAlias_Success(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	alias := "alias1 #public"
	finalAlias := "alias1"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition(finalAlias))

	d := b.GetDefinition(key)
	a := b.GetDefinition(finalAlias)
	assert.Equal(t, d.Factory(&container{}), a.Factory(&container{}))
	assert.Equal(t, d, a.AliasOf)
	assert.NotEqual(t, d.Tags.All(), a.Tags.All())
}

func TestContainerBuilder_SetAlias_HasOwnTags(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	taggedAlias := "alias1 #tag1 #private"
	expectedAlias := "alias1"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(taggedAlias, key)
	assert.False(t, b.HasDefinition(taggedAlias))
	assert.True(t, b.HasDefinition(expectedAlias))

	d := b.GetDefinition(key)
	a := b.GetDefinition(expectedAlias)

	assert.Equal(t, d.Factory(&container{}), a.Factory(&container{}))
	assert.Equal(t, d, a.AliasOf)
	assert.True(t, a.Tags.Has("tag1"))
	assert.True(t, a.Tags.Has("private"))

}

func TestContainerBuilder_AddProvider(t *testing.T) {
	p1 := ProviderFunc(func(_ ContainerBuilder) {})
	p2 := ProviderFunc(func(_ ContainerBuilder) {})

	b := NewContainerBuilder()

	b.AddProvider([]Provider{p1, p2})
	assert.Equal(t, 2, len(b.providers))
}

func TestContainerBuilder_AddProvider_PanicsIfResolved(t *testing.T) {
	p1 := ProviderFunc(func(_ ContainerBuilder) {})
	p2 := ProviderFunc(func(_ ContainerBuilder) {})

	b := NewContainerBuilder()
	b.resolved = true

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.AddProvider([]Provider{p1, p2})
	})
}

func TestContainerBuilder_AddResolver(t *testing.T) {
	p1 := ResolverFunc(func(_ ContainerBuilder) {})
	p2 := ResolverFunc(func(_ ContainerBuilder) {})

	b := NewContainerBuilder()

	b.AddResolver([]Resolver{p1, p2})
	assert.Equal(t, 2, len(b.resolvers))
}

func TestContainerBuilder_AddResolver_PanicsIfResolved(t *testing.T) {
	p1 := ResolverFunc(func(_ ContainerBuilder) {})
	p2 := ResolverFunc(func(_ ContainerBuilder) {})

	b := NewContainerBuilder()
	b.resolved = true

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.AddResolver([]Resolver{p1, p2})
	})
}

func TestContainerBuilder_GetContainer(t *testing.T) {
	spyProvide := new(bool)
	spyResolve := new(bool)
	p1 := ProviderFunc(func(_ ContainerBuilder) {
		*spyProvide = true
	})
	p2 := ResolverFunc(func(_ ContainerBuilder) {
		*spyResolve = true
	})

	b := NewContainerBuilder()
	b.AddProvider([]Provider{p1})
	b.AddResolver([]Resolver{p2})
	c := b.GetContainer()

	assert.True(t, b.resolved)
	assert.True(t, *spyProvide)
	assert.True(t, *spyResolve)
	assert.Same(t, b, c.builder)
	assert.Empty(t, c.instances.All())
}

func TestContainerBuilder_GetContainer_UsesSameResolvedBuilder(t *testing.T) {
	p1 := ProviderFunc(func(_ ContainerBuilder) {})

	b := NewContainerBuilder()
	b.AddProvider([]Provider{p1})
	c1 := b.GetContainer()
	c2 := b.GetContainer()

	assert.Same(t, b, c1.builder)
	assert.Same(t, c1.builder, c2.builder)
	assert.NotSame(t, c1, c2)
}

func TestContainerBuilder_GetTaggedKeys(t *testing.T) {
	key1 := "key1 #tag=one #"
	key2 := "key2 #tag=two"
	key3 := "key3 #other"
	val1 := func(_ Container) interface{} { return 1 }
	val2 := func(_ Container) interface{} { return 2 }
	val3 := func(_ Container) interface{} { return 3 }

	b := NewContainerBuilder()
	b.SetDefinition(key1, val1)
	b.SetDefinition(key2, val2)
	b.SetDefinition(key3, val3)

	ds := b.GetTaggedKeys("tag", []string{})
	assert.Subset(t, []string{"key1", "key2"}, ds)
	assert.Len(t, ds, 2)

	ds = b.GetTaggedKeys("tag", []string{"one", "two"})
	assert.Subset(t, []string{"key1", "key2"}, ds)
	assert.Len(t, ds, 2)

	ds = b.GetTaggedKeys("tag", []string{"one"})
	assert.Equal(t, []string{"key1"}, ds)

	ds = b.GetTaggedKeys("tag", []string{"two"})
	assert.Equal(t, []string{"key2"}, ds)
}

func TestContainerBuilder_GetContainer_IfConcurrent(t *testing.T) {
	init := new(int)
	*init = 0

	p := ProviderFunc(func(b ContainerBuilder) {
		b.SetDefinition("s1", func(cb Container) interface{} {
			return false
		})

		if *init == 1 {
			b.SetDefinition("s1", func(cb Container) interface{} {
				return true
			})
		}
		*init++
	})

	b := NewContainerBuilder()
	b.AddProvider([]Provider{p})

	for i := 0; i < 1000; i++ {
		go func() {
			c := b.GetContainer()
			assert.False(t, c.Get("s1").(bool))
		}()
	}
}

func TestContainerBuilder_SetMany_setsAllTypes(t *testing.T) {
	b := NewContainerBuilder()
	b.SetMany([]Some{
		{Key: "service #private", Val: func(c Container) interface{} {
			return c.GetParameter("param").(int)
		}},
		{Key: "param", Val: 1},
		{Key: "alias", Val: "service"},
		{Key: "injectable #inject", Val: struct{}{}},
	}...)

	assert.True(t, b.HasDefinition("service"))
	assert.True(t, b.HasDefinition("injectable"))
	assert.True(t, b.HasParameter("param"))
	assert.True(t, b.HasDefinition("alias"))
}

func TestContainerBuilder_SetInjectable(t *testing.T) {
	type Empty struct{}
	type NoneInject struct {
		F1 string
	}
	type ServiceInject struct {
		F1 string `inject:"s1"`
	}
	type ParameterInject struct {
		P1 string `inject:"_p1"`
	}
	type UnexportedField struct {
		f1 string `inject:"s1"`
	}
	type EmptyInjectKey struct {
		F1 string `inject:""`
	}
	type Composed struct {
		S1 string          `inject:"s1"`
		P1 string          `inject:"_p1"`
		S2 ServiceInject   `inject:"s2"`
		P2 ParameterInject `inject:"p2"`
	}

	for _, data := range []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"can build empty struct", Empty{}, Empty{}},
		{"can build struct without inject tags", NoneInject{}, NoneInject{}},
		{"can build struct injecting services in exported fields", ServiceInject{}, ServiceInject{F1: "bye!"}},
		{"can build struct injecting parameters in exported fields", ParameterInject{}, ParameterInject{P1: "hi!"}},
		{"can build pointers to struct injecting services in exported fields", &ServiceInject{}, &ServiceInject{F1: "bye!"}},
	} {
		t.Run(data.name, func(t *testing.T) {
			b := NewContainerBuilder()
			b.SetParameter("p1", "hi!")
			b.SetDefinition("s1", func(c Container) interface{} { return "bye!" })

			b.SetInjectable("i1", data.value)
			c := b.GetContainer()
			s := c.Get("i1")

			assert.Equal(t, data.expected, s)
		})
	}

	for _, data := range []struct {
		name  string
		value interface{}
	}{
		{"panics if unexported field to inject", UnexportedField{f1: ""}},
		{"panics if empty injection key", EmptyInjectKey{}},
	} {
		t.Run(data.name, func(t *testing.T) {
			b := NewContainerBuilder()
			b.SetParameter("p1", "hi!")
			b.SetDefinition("s1", func(c Container) interface{} { return "bye!" })

			assert.Panics(t, func() {
				b.SetInjectable("i1", data.value)
			})
		})
	}

	t.Run("can build composed structs", func(t *testing.T) {
		b := NewContainerBuilder()
		b.SetParameter("p1", "hi!")
		b.SetDefinition("s1", func(c Container) interface{} { return "bye!" })
		b.SetInjectable("s2", ServiceInject{})
		b.SetInjectable("p2", ParameterInject{})

		b.SetInjectable("c1", Composed{})
		c := b.GetContainer()
		s := c.Get("c1")

		e := Composed{
			S1: "bye!",
			P1: "hi!",
			S2: ServiceInject{F1: "bye!"},
			P2: ParameterInject{P1: "hi!"},
		}

		assert.Equal(t, e, s)
	})

}
