// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var dummyProvider = ProviderFunc(func(ContainerBuilder) {})
var dummyResolver = ResolverFunc(func(ContainerBuilder) {})

func TestNewContainerBuilder(t *testing.T) {
	t.Run("returns an initialized builder", func(t *testing.T) {
		b := NewContainerBuilder()
		assert.NotNil(t, b.definitions)
		assert.NotNil(t, b.providers)
		assert.Len(t, b.definitions, 0)
		assert.Len(t, b.providers, 0)
		assert.False(t, b.resolved)
	})
}

func testSetMethodsCommon(t *testing.T, kindTag string, f func(b ContainerBuilder, key string, tags ...map[string]string)) {
	b := NewContainerBuilder()
	b.SetValue("a1", "a1")

	t.Run("adds definition without tags", func(t *testing.T) {
		f(b, "k1")
		assert.True(t, b.HasDefinition("k1"))
		assert.True(t, b.GetDefinition("k1").HasTag(kindTag))
		assert.Equal(t, 1, len(b.GetDefinition("k1").Tags))
	})

	t.Run("adds definition with tags", func(t *testing.T) {
		f(b, "k2", map[string]string{"t2": "v2"})
		assert.True(t, b.HasDefinition("k2"))
		assert.True(t, b.GetDefinition("k2").HasTag("t2"))
		assert.Equal(t, "v2", b.GetDefinition("k2").GetTag("t2"))
	})

	t.Run("adds definition tags in key", func(t *testing.T) {
		f(b, "k3 #t3=v3")
		assert.False(t, b.HasDefinition("k3 #t3=v3"))
		assert.True(t, b.HasDefinition("k3"))
		assert.True(t, b.GetDefinition("k3").HasTag("t3"))
		assert.Equal(t, "v3", b.GetDefinition("k3").GetTag("t3"))
	})

	t.Run("overwrites tags in key", func(t *testing.T) {
		f(b, "k4 #t4=v4", map[string]string{"t4": "v44", "t5": "v5"})
		assert.False(t, b.HasDefinition("k4 #t4=v4"))
		assert.True(t, b.HasDefinition("k4"))
		assert.True(t, b.GetDefinition("k4").HasTag("t4"))
		assert.Equal(t, "v44", b.GetDefinition("k4").GetTag("t4"))
		assert.True(t, b.GetDefinition("k4").HasTag("t5"))
		assert.Equal(t, "v5", b.GetDefinition("k4").GetTag("t5"))
	})

	t.Run(fmt.Sprintf("manages %s tag", kindTag), func(t *testing.T) {
		f(b, "k7")
		f(b, "k8", map[string]string{kindTag: "my-value"})

		assert.Equal(t, "", b.GetDefinition("k7").GetTag(kindTag))
		assert.Equal(t, "my-value", b.GetDefinition("k8").GetTag(kindTag))
	})

	t.Run("panics if resolved", func(t *testing.T) {
		b.GetContainer()
		assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
			f(b, "key")
		})
	})
}

func testSetMethodsReplaceAlias(t *testing.T, f func(b ContainerBuilder, key string)) {
	t.Run("replaces alias", func(t *testing.T) {
		key := "Key"
		alias := "alias"
		c := &container{}

		b := NewContainerBuilder()
		b.SetValue(key, "aliased")
		b.SetAlias(alias, key)
		a := b.GetDefinition(alias)
		k := b.GetDefinition(key)
		assert.EqualValues(t, a.Factory(c), k.Factory(c))
		assert.Equal(t, k, a.AliasOf)

		f(b, alias)
		assert.True(t, b.HasDefinition(alias))

		a = b.GetDefinition(alias)
		k = b.GetDefinition(key)
		assert.NotEqualValues(t, a.Factory(c), k.Factory(c))
		assert.Nil(t, a.AliasOf)
	})
}

func TestContainerBuilder_SetValue(t *testing.T) {
	testSetMethodsCommon(t, TagValue, func(b ContainerBuilder, key string, tags ...map[string]string) {
		b.SetValue(key, "dummy", tags...)
	})

	testSetMethodsReplaceAlias(t, func(b ContainerBuilder, key string) {
		b.SetValue(key, "dummy")
	})

	t.Run("binds values of type", func(t *testing.T) {
		data := map[string]interface{}{
			"boolean": true,
			"string":  "abc",
			"integer": 123,
			"float":   1.23,
			"slice":   []byte{'1', '2', '3'},
			"map":     map[string]int16{"1": 1, "2": 2, "3": 3},
			"struct": struct {
				Greet  string
				Number int
			}{
				"hello",
				123,
			},
			"pointer": NewContainerBuilder(),
		}

		for key, val := range data {
			t.Run(key, func(t *testing.T) {
				b := NewContainerBuilder()
				b.SetValue(key, val)
				c := b.GetContainer()

				assert.True(t, b.HasDefinition(key))
				assert.Equal(t, val, b.GetDefinition(key).Factory(c))
			})
		}
	})
}

func TestContainerBuilder_SetFactory(t *testing.T) {
	testSetMethodsCommon(t, TagFactory, func(b ContainerBuilder, key string, tags ...map[string]string) {
		b.SetFactory(key, dummyFactory, tags...)
	})

	testSetMethodsReplaceAlias(t, func(b ContainerBuilder, key string) {
		b.SetValue(key, dummyFactory)
	})
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
		P1 string `inject:"p1"`
	}
	type UnexportedField struct {
		f1 string `inject:"s1"`
	}
	type EmptyInjectKey struct {
		F1 string `inject:""`
	}
	type Composed struct {
		S1 string          `inject:"s1"`
		P1 string          `inject:"p1"`
		S2 ServiceInject   `inject:"s2"`
		P2 ParameterInject `inject:"p2"`
	}

	testSetMethodsCommon(t, TagInject, func(b ContainerBuilder, key string, tags ...map[string]string) {
		b.SetInjectable(key, Empty{}, tags...)
	})

	testSetMethodsReplaceAlias(t, func(b ContainerBuilder, key string) {
		b.SetInjectable(key, Empty{})
	})

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
			b.SetValue("p1", "hi!")
			b.SetFactory("s1", func(c Container) interface{} { return "bye!" })

			b.SetInjectable("i1", data.value)
			c := b.GetContainer()
			s := c.Get("i1")

			assert.Equal(t, data.expected, s)
		})
	}

	for _, data := range []struct {
		name  string
		value interface{}
		error string
	}{
		{"panics if not a struct", "dummy", "invalid injectable for key i1, only structs can be injectables"},
		{"panics if unexported field to inject", UnexportedField{f1: ""}, "unexported field github.com/golossus/di/f1 can not be injected"},
		{"panics if empty injection key", EmptyInjectKey{}, "no injection key present for field EmptyInjectKey: F1"},
	} {
		t.Run(data.name, func(t *testing.T) {
			b := NewContainerBuilder()
			b.SetValue("p1", "hi!")
			b.SetFactory("s1", func(c Container) interface{} { return "bye!" })

			assert.PanicsWithValue(t, data.error, func() {
				b.SetInjectable("i1", data.value)
			})
		})
	}

	t.Run("can build composed structs", func(t *testing.T) {
		b := NewContainerBuilder()
		b.SetValue("p1", "hi!")
		b.SetFactory("s1", func(c Container) interface{} { return "bye!" })
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

func TestContainerBuilder_SetAlias(t *testing.T) {
	testSetMethodsCommon(t, TagAlias, func(b ContainerBuilder, key string, tags ...map[string]string) {
		b.SetAlias(key, "a1", tags...)
	})

	b := NewContainerBuilder()
	t.Run("aliases a service", func(t *testing.T) {
		b.SetFactory("key", dummyFactory)
		assert.PanicsWithValue(t, "definition with id 'key' already exists and alias cannot be set", func() {
			b.SetAlias("key", "key")
		})
	})

	t.Run("panics if service does not exist", func(t *testing.T) {
		assert.PanicsWithValue(t, "definition with id 'def' does not exist and alias cannot be set", func() {
			b.SetAlias("key2", "def")
		})
	})

	t.Run("panics if service with same key already exist", func(t *testing.T) {
		b.SetFactory("key", dummyFactory)
		assert.PanicsWithValue(t, "definition with id 'key' already exists and alias cannot be set", func() {
			b.SetAlias("key", "key")
		})
	})
}

func TestContainerBuilder_SetAll(t *testing.T) {
	t.Run("binds all kinds of definitions", func(t *testing.T) {
		b := NewContainerBuilder()
		b.SetAll([]Binding{
			{Key: "service #factory", Target: func(c Container) interface{} {
				return c.Get("param").(int)
			}},
			{Key: "param #value", Target: 1},
			{Key: "alias #alias", Target: "service"},
			{Key: "injectable #inject", Target: struct{}{}},
			{Key: "service2", Target: func(c Container) interface{} {
				return c.Get("param").(int)
			}},
		}...)

		assert.True(t, b.HasDefinition("service"))
		assert.True(t, b.HasDefinition("injectable"))
		assert.True(t, b.HasDefinition("param"))
		assert.True(t, b.HasDefinition("alias"))
		assert.True(t, b.HasDefinition("service2"))
		assert.Equal(t, "factory", b.GetDefinition("service").Kind)
		assert.Equal(t, "inject", b.GetDefinition("injectable").Kind)
		assert.Equal(t, "value", b.GetDefinition("param").Kind)
		assert.Equal(t, "alias", b.GetDefinition("alias").Kind)
		assert.Equal(t, "factory", b.GetDefinition("service2").Kind)
	})

	t.Run("panics", func(t *testing.T) {
		sharedData := []struct {
			name   string
			key    string
			target interface{}
			error  string
		}{
			{"if invalid #priority=abc", "dummy #priority=abc", dummyFactory, "priority tag value 'abc' is not a valid number for key 'dummy'"},
			{"if invalid #private=off", "dummy #private=off", dummyFactory, "private tag value 'off' is not a valid boolean for key 'dummy'"},
			{"if invalid #shared=on", "dummy #shared=on", dummyFactory, "shared tag value 'on' is not a valid boolean for key 'dummy'"},
			{"if overlapping kinds", "dummy #factory #value", dummyFactory, "tag 'value' can't be used simultaneously with [factory value alias inject] for key 'dummy'"},
		}

		b := NewContainerBuilder()
		for _, data := range sharedData {
			t.Run(data.name, func(t *testing.T) {
				assert.PanicsWithValue(t, data.error, func() {
					b.SetAll(Binding{Key: data.key, Target: data.target})
				})
			})
		}
	})

	t.Run("panics if invalid factory", func(t *testing.T) {
		b := NewContainerBuilder()
		assert.Panics(t, func() {
			b.SetAll(Binding{Key: "#factory", Target: 1})
		})
	})
}

func TestContainerBuilder_GetTaggedKeys(t *testing.T) {
	key1 := "key1 #tag=one"
	key2 := "key2 #tag=two #priority=10"
	key3 := "key3 #other"

	b := NewContainerBuilder()
	b.SetFactory(key1, dummyFactory)
	b.SetFactory(key2, dummyFactory)
	b.SetFactory(key3, dummyFactory)

	t.Run("get keys by tag", func(t *testing.T) {
		ds := b.GetTaggedKeys("tag", []string{})
		assert.Subset(t, []string{"key2", "key1"}, ds)
		assert.Len(t, ds, 2)
	})

	t.Run("get keys by tag and all values", func(t *testing.T) {
		ds := b.GetTaggedKeys("tag", []string{"one", "two"})
		assert.Subset(t, []string{"key2", "key1"}, ds)
		assert.Len(t, ds, 2)
	})

	t.Run("get keys by tag and constrained values", func(t *testing.T) {
		ds := b.GetTaggedKeys("tag", []string{"one"})
		assert.Subset(t, []string{"key1"}, ds)
		assert.Len(t, ds, 1)

		ds = b.GetTaggedKeys("tag", []string{"two"})
		assert.Subset(t, []string{"key2"}, ds)
		assert.Len(t, ds, 1)
	})
}

func TestContainerBuilder_AddProvider(t *testing.T) {
	t.Run("adds providers", func(t *testing.T) {
		b := NewContainerBuilder()
		b.AddProvider([]Provider{dummyProvider, dummyProvider})

		assert.Equal(t, 2, len(b.providers))
	})

	t.Run("panics if resolved", func(t *testing.T) {
		b := NewContainerBuilder()
		b.GetContainer()

		assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
			b.AddProvider([]Provider{dummyProvider, dummyProvider})
		})
	})
}

func TestContainerBuilder_AddResolver(t *testing.T) {
	t.Run("adds providers", func(t *testing.T) {
		b := NewContainerBuilder()
		b.AddResolver([]Resolver{dummyResolver, dummyResolver})

		assert.Equal(t, 2, len(b.resolvers))
	})

	t.Run("panics if resolved", func(t *testing.T) {
		b := NewContainerBuilder()
		b.GetContainer()

		assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
			b.AddResolver([]Resolver{dummyResolver, dummyResolver})
		})
	})
}

func TestContainerBuilder_GetContainer(t *testing.T) {
	t.Run("resolves builder services", func(t *testing.T) {
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
		assert.Empty(t, c.instances)
	})

	t.Run("reuses resolved builder", func(t *testing.T) {
		p1 := ProviderFunc(func(_ ContainerBuilder) {})

		b := NewContainerBuilder()
		b.AddProvider([]Provider{p1})
		c1 := b.GetContainer()
		c2 := b.GetContainer()

		assert.Same(t, b, c1.builder)
		assert.Same(t, c1.builder, c2.builder)
		assert.NotSame(t, c1, c2)
	})

	t.Run("is concurrently safe", func(t *testing.T) {
		init := new(int)
		*init = 0

		p := ProviderFunc(func(b ContainerBuilder) {
			b.SetFactory("s1", func(cb Container) interface{} {
				return false
			})

			if *init == 1 {
				b.SetFactory("s1", func(cb Container) interface{} {
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
	})
}
