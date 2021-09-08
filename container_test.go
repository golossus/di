package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainer_Get(t *testing.T) {
	newA := func(cb *containerBuilder) int {
		return 1
	}

	b := NewContainerBuilder()
	b.SetDefinition("a", newA)
	c := b.GetContainer()

	r := c.Get("a").(int)

	assert.Equal(t, 1, r)
}

func TestContainer_Get_SharedService(t *testing.T) {
	spy := 0
	newA := func(cb *containerBuilder) int {
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
