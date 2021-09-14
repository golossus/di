// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestItemHash_All_EmptyIfNew(t *testing.T) {
	hash := newItemHash()
	assert.Equal(t, 0, len(hash.All()))
}

func TestItemHash_Set_AddsNewItem(t *testing.T) {
	key1 := "key1"
	val1 := "value1"

	hash := newItemHash()
	assert.False(t, hash.Has(key1))

	hash.set(key1, val1)
	assert.True(t, hash.Has(key1))
	assert.Equal(t, val1, hash.Get(key1))
}

func TestItemHash_Set_AddsMoreItems(t *testing.T) {
	key1 := "key1"
	val1 := "value1"
	key2 := "key2"
	val2 := "value2"

	hash := newItemHash()
	hash.set(key1, val1)
	hash.set(key2, val2)

	assert.True(t, hash.Has(key1))
	assert.True(t, hash.Has(key2))
	assert.Equal(t, val1, hash.Get(key1))
	assert.Equal(t, val2, hash.Get(key2))
}

func TestItemHash_Set_ReplacesItemsWithSameKey(t *testing.T) {
	key1 := "key1"
	val1 := "value1"
	key2 := "key2"
	val2 := "value2"

	hash := newItemHash()
	hash.set(key1, val1)
	hash.set(key2, val2)
	hash.set(key1, val2)

	assert.True(t, hash.Has(key1))
	assert.True(t, hash.Has(key2))
	assert.Equal(t, val2, hash.Get(key1))
	assert.Equal(t, val2, hash.Get(key2))
}

func TestItemHash_Del_RemovesItems(t *testing.T) {
	key1 := "key1"
	val1 := "value1"
	key2 := "key2"
	val2 := "value2"

	hash := newItemHash()
	hash.set(key1, val1)
	hash.set(key2, val2)
	hash.del(key1, key2)

	assert.False(t, hash.Has(key1))
	assert.False(t, hash.Has(key2))
	assert.Equal(t, 0, len(hash.All()))
}

func TestItemHash_Get_PanicsIfItemNotFound(t *testing.T) {
	key1 := "key1"

	hash := newItemHash()

	assert.PanicsWithValue(t, "item with Key 'key1' not found", func() {
		hash.Get(key1)
	})
}