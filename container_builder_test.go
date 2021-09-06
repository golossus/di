package di

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestContainerBuilder_SetDefinition_Success(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1"
	val := func(b *ContainerBuilder) int {
		return 1
	}

	b.SetDefinition(key, val)
	assert.True(t, b.HasDefinition(key))

	f, ok := b.GetDefinition(key).Build.(func(*ContainerBuilder) int)
	assert.True(t, ok)
	assert.Equal(t, val(&ContainerBuilder{}), f(&ContainerBuilder{}))
}

func TestContainerBuilder_SetDefinition_RemovesAlias(t *testing.T) {
	b := NewContainerBuilder()
	alias := "alias"
	key := "key"
	val := func(b *ContainerBuilder) int {
		return 1
	}

	b.SetDefinition(key, val)
	b.SetAlias(alias, key)
	assert.True(t, b.HasAlias(alias))
	assert.True(t, b.HasDefinition(alias))

	f, ok := b.GetDefinition(alias).Build.(func(*ContainerBuilder) int)
	assert.True(t, ok)
	assert.Equal(t, val(&ContainerBuilder{}), f(&ContainerBuilder{}))

	b.SetDefinition(alias, val)
	assert.True(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition(alias))
	assert.False(t, b.HasAlias(alias))

	f, ok = b.GetDefinition(key).Build.(func(*ContainerBuilder) int)
	assert.True(t, ok)
	assert.Equal(t, val(&ContainerBuilder{}), f(&ContainerBuilder{}))
}

func TestContainerBuilder_SetDefinition_SuccessWithTags(t *testing.T) {
	b := NewContainerBuilder()
	key := "key1 #private #shared #other"
	val := func(b *ContainerBuilder) int {
		return 1
	}

	b.SetDefinition(key, val)
	assert.False(t, b.HasDefinition(key))
	assert.True(t, b.HasDefinition("key1"))


	d := b.GetDefinition("key1")
	assert.True(t, d.Private())
	assert.True(t, d.Shared())
	assert.True(t, d.Tags.has("other"))

	f, ok := d.Build.(func(*ContainerBuilder) int)
	assert.True(t, ok)
	assert.Equal(t, val(&ContainerBuilder{}), f(&ContainerBuilder{}))
}
