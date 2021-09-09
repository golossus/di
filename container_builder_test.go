package di

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewContainerBuilder_ReturnsInitialised(t *testing.T) {
	b := NewContainerBuilder()
	assert.NotNil(t, b.params)
	assert.NotNil(t, b.defs)
	assert.NotNil(t, b.alias)
	assert.NotNil(t, b.providers)
	assert.NotNil(t, b.parser)
	assert.Len(t, b.params.all(), 0)
	assert.Len(t, b.defs.all(), 0)
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
	val := func(b ContainerInterface) int {
		return 1
	}

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.SetDefinition(key, val)
	})
}

func TestContainerBuilder_SetDefinition_PanicsIfInvalidConstructor(t *testing.T) {
	b := NewContainerBuilder()

	key := "key1"
	val := 1

	assert.PanicsWithValue(t, "invalid constructor kind 'int', must be a function", func() {
		b.SetDefinition(key, val)
	})
}

func TestContainerBuilder_SetDefinition_PanicsIfInvalidConstructorReturn(t *testing.T) {
	b := NewContainerBuilder()

	key := "key1"
	val := func(b int) {}
	val2 := func(b int) (int, int) {
		return 1, 1
	}

	assert.PanicsWithValue(t, "constructor 'func(int)' should return a single value", func() {
		b.SetDefinition(key, val)
	})
	assert.PanicsWithValue(t, "constructor 'func(int) (int, int)' should return a single value", func() {
		b.SetDefinition(key, val2)
	})
}

func TestContainerBuilder_SetDefinition_PanicsIfInvalidConstructorArguments(t *testing.T) {
	b := NewContainerBuilder()

	key := "key1"
	val := func(b int) int {
		return 1
	}

	assert.PanicsWithValue(t, "constructor 'func(int) int' can only receive a 'ContainerInterface' argument", func() {
		b.SetDefinition(key, val)
	})
}

func TestContainerBuilder_SetDefinition_Success(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	val := func(b ContainerInterface) int {
		return 1
	}

	b.SetDefinition(key, val)
	assert.True(t, b.HasDefinition(key))

	f, ok := b.GetDefinition(key).Build.(func(ContainerInterface) int)
	assert.True(t, ok)
	assert.Equal(t, val(&container{}), f(&container{}))
}

func TestContainerBuilder_SetDefinition_SuccessIfConstructorHasNoArgument(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	val := func() int {
		return 1
	}

	b.SetDefinition(key, val)
	assert.True(t, b.HasDefinition(key))

	f, ok := b.GetDefinition(key).Build.(func() int)
	assert.True(t, ok)
	assert.Equal(t, val(), f())
}

func TestContainerBuilder_SetDefinition_RemovesAlias(t *testing.T) {
	b := NewContainerBuilder()
	alias := "alias"
	key := "key"
	val := func(b ContainerInterface) int {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasAlias(alias))
	assert.True(t, b.HasDefinition(alias))

	f, ok := b.GetDefinition(alias).Build.(func( ContainerInterface) int)
	assert.True(t, ok)
	assert.Equal(t, val(&container{}), f(&container{}))

	b.SetDefinition(alias, val)
	assert.True(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition(alias))
	assert.False(t, b.HasAlias(alias))

	f, ok = b.GetDefinition(key).Build.(func( ContainerInterface) int)
	assert.True(t, ok)
	assert.Equal(t, val(&container{}), f(&container{}))
}

func TestContainerBuilder_SetDefinition_SuccessWithTags(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1 #private #shared #other"
	val := func(b ContainerInterface) int {
		return 1
	}

	b.SetDefinition(key, val)
	assert.False(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition("key1"))

	d := b.GetDefinition("key1")
	assert.True(t, d.Private())
	assert.True(t, d.Shared())
	assert.True(t, d.Tags.has("other"))

	f, ok := d.Build.(func(ContainerInterface) int)
	assert.True(t, ok)
	assert.Equal(t, val(&container{}), f(&container{}))
}

func TestContainerBuilder_SetAlias_PanicsIfResolved(t *testing.T) {
	b := NewContainerBuilder()
	b.resolved = true

	key := "key1"
	key2 := "key2"

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.SetDefinition(key, key2)
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
	val := func(b ContainerInterface) int {
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
	val := func(b ContainerInterface) int {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition(alias))
	assert.True(t, b.HasAlias(alias))
	assert.False(t, b.HasAlias(key))
	assert.Equal(t, b.GetAlias(alias), key)

	d, okDef := b.GetDefinition(key).Build.(func(ContainerInterface) int)
	a, okAlias := b.GetDefinition(alias).Build.(func(ContainerInterface) int)
	assert.True(t, okDef)
	assert.True(t, okAlias)
	assert.Equal(t, d(&container{}), a(&container{}))
}

func TestContainerBuilder_SetAlias_IgnoresKeyTags(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	taggedAlias := "alias1 #tag1"
	expectedAlias := "alias1"
	val := func(b ContainerInterface) int {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(taggedAlias, key)
	assert.False(t, b.HasAlias(taggedAlias))
	assert.True(t, b.HasAlias(expectedAlias))
	assert.False(t, b.HasDefinition(taggedAlias))
	assert.True(t, b.HasDefinition(expectedAlias))
	assert.Equal(t, b.GetAlias(expectedAlias), key)

	d, okDef := b.GetDefinition(key).Build.(func(ContainerInterface) int)
	a, okAlias := b.GetDefinition(expectedAlias).Build.(func(ContainerInterface) int)
	assert.True(t, okDef)
	assert.True(t, okAlias)
	assert.Equal(t, d(&container{}), a(&container{}))
}

type DummyProvider struct {
	spyRegistered bool
	spyResolved   bool
}

func (p *DummyProvider) Register(b ContainerBuilderInterface) {
	p.spyRegistered = true
}

func (p *DummyProvider) Resolve(b ContainerBuilderInterface) {
	p.spyResolved = true
}

func TestContainerBuilder_AddProviders(t *testing.T) {
	p1 := &DummyProvider{}
	p2 := &DummyProvider{}

	b := NewContainerBuilder()

	b.AddProviders()
	assert.Equal(t, 0, len(b.providers))

	b.AddProviders(p1, p2)
	assert.Equal(t, 2, len(b.providers))
}

func TestContainerBuilder_AddProviders_PanicsIfResolved(t *testing.T) {
	p1 := &DummyProvider{}
	p2 := &DummyProvider{}

	b := NewContainerBuilder()
	b.resolved = true

	b.AddProviders()
	assert.Equal(t, 0, len(b.providers))

	assert.PanicsWithValue(t, "container is resolved and new items can not be set", func() {
		b.AddProviders(p1, p2)
	})
}

func TestContainerBuilder_GetContainer(t *testing.T) {
	p1 := &DummyProvider{}
	p2 := &DummyProvider{}

	b := NewContainerBuilder()
	b.AddProviders(p1, p2)
	c := b.GetContainer()

	assert.True(t, b.resolved)
	assert.True(t, p1.spyRegistered)
	assert.True(t, p1.spyResolved)
	assert.True(t, p2.spyRegistered)
	assert.True(t, p2.spyResolved)
	assert.Same(t, b, c.builder)
	assert.Empty(t, c.instances.all())
}

func TestContainerBuilder_GetContainer_UsesSameResolvedBuilder(t *testing.T) {
	p1 := &DummyProvider{}
	p2 := &DummyProvider{}

	b := NewContainerBuilder()
	b.AddProviders(p1, p2)
	c1 := b.GetContainer()
	c2 := b.GetContainer()

	assert.Same(t, b, c1.builder)
	assert.Same(t, c1.builder, c2.builder)
	assert.NotSame(t, c1, c2)
}
