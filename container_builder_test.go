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
	assert.NotNil(t, b.alias)
	assert.NotNil(t, b.providers)
	assert.NotNil(t, b.parser)
	assert.Len(t, b.parameters.all(), 0)
	assert.Len(t, b.definitions.all(), 0)
	assert.Len(t, b.alias.all(), 0)
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
	key := "some-key-with #tag"
	keyParsed := "some-key-with"
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
	key := "key"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasAlias(alias))
	assert.True(t, b.HasDefinition(alias))

	f := b.GetDefinition(alias).Factory
	assert.Equal(t, val(&container{}), f(&container{}))

	b.SetDefinition(alias, val)
	assert.True(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition(alias))
	assert.False(t, b.HasAlias(alias))

	f = b.GetDefinition(key).Factory
	assert.Equal(t, val(&container{}), f(&container{}))
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
	assert.True(t, d.Tags.has("other"))

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
	alias := "alias1"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition(alias))
	assert.True(t, b.HasAlias(alias))
	assert.False(t, b.HasAlias(key))
	assert.Equal(t, b.GetAlias(alias), key)

	d := b.GetDefinition(key).Factory
	a := b.GetDefinition(alias).Factory
	assert.Equal(t, d(&container{}), a(&container{}))
}

func TestContainerBuilder_SetAlias_IgnoresKeyTags(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	taggedAlias := "alias1 #tag1"
	expectedAlias := "alias1"
	val := func(c Container) interface{} {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(taggedAlias, key)
	assert.False(t, b.HasAlias(taggedAlias))
	assert.True(t, b.HasAlias(expectedAlias))
	assert.False(t, b.HasDefinition(taggedAlias))
	assert.True(t, b.HasDefinition(expectedAlias))
	assert.Equal(t, b.GetAlias(expectedAlias), key)

	d := b.GetDefinition(key).Factory
	a := b.GetDefinition(expectedAlias).Factory
	assert.Equal(t, d(&container{}), a(&container{}))
}

type DummyProviderResolver struct {
	spyProvide, spyResolve bool
}

func (p *DummyProviderResolver) Provide(_ ContainerBuilder) { p.spyProvide = true }
func (p *DummyProviderResolver) Resolve(_ ContainerBuilder) { p.spyResolve = true }

func TestContainerBuilder_AddProvider(t *testing.T) {
	p1 := &DummyProviderResolver{}
	p2 := &DummyProviderResolver{}

	b := NewContainerBuilder()

	b.AddProvider([]Provider{p1, p2})
	assert.Equal(t, 2, len(b.providers))
}

func TestContainerBuilder_AddProvider_PanicsIfResolved(t *testing.T) {
	p1 := &DummyProviderResolver{}
	p2 := &DummyProviderResolver{}

	b := NewContainerBuilder()
	b.resolved = true

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.AddProvider([]Provider{p1, p2})
	})
}

func TestContainerBuilder_AddResolver(t *testing.T) {
	p1 := &DummyProviderResolver{}
	p2 := &DummyProviderResolver{}

	b := NewContainerBuilder()

	b.AddResolver([]Resolver{p1, p2})
	assert.Equal(t, 2, len(b.resolvers))
}

func TestContainerBuilder_AddResolver_PanicsIfResolved(t *testing.T) {
	p1 := &DummyProviderResolver{}
	p2 := &DummyProviderResolver{}

	b := NewContainerBuilder()
	b.resolved = true

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.AddResolver([]Resolver{p1, p2})
	})
}

func TestContainerBuilder_GetContainer(t *testing.T) {
	p1 := &DummyProviderResolver{}
	p2 := &DummyProviderResolver{}

	b := NewContainerBuilder()
	b.AddProvider([]Provider{p1, p2})
	b.AddResolver([]Resolver{p1, p2})
	c := b.GetContainer()

	assert.True(t, b.resolved)
	assert.True(t, p1.spyProvide)
	assert.True(t, p1.spyResolve)
	assert.True(t, p2.spyProvide)
	assert.True(t, p2.spyResolve)
	assert.Same(t, b, c.builder)
	assert.Empty(t, c.instances.all())
}

func TestContainerBuilder_GetContainer_UsesSameResolvedBuilder(t *testing.T) {
	p1 := &DummyProviderResolver{}
	p2 := &DummyProviderResolver{}

	b := NewContainerBuilder()
	b.AddProvider([]Provider{p1, p2})
	c1 := b.GetContainer()
	c2 := b.GetContainer()

	assert.Same(t, b, c1.builder)
	assert.Same(t, c1.builder, c2.builder)
	assert.NotSame(t, c1, c2)
}

func TestContainerBuilder_GetDefinitionTaggedBy(t *testing.T) {
	key1 := "key1 #tag=one"
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
