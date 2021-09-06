package di

import "fmt"

type itemHash struct {
	hash map[string]interface{}
}

func newItemHash() *itemHash {
	return &itemHash{
		hash: make(map[string]interface{}),
	}
}

func (i *itemHash) set(key string, val interface{}) {
	i.hash[key] = val
}

func (i *itemHash) has(key string) bool {
	_, ok := i.hash[key]
	return ok
}

func (i *itemHash) get(key string) interface{} {
	val, ok := i.hash[key]
	if !ok {
		panic(fmt.Sprintf("item with key '%s' not found", key))
	}
	return val
}

func (i *itemHash) del(keys ...string) {
	for _, key := range keys {
		delete(i.hash, key)
	}
}

func (i *itemHash) all() map[string]interface{} {
	all := make(map[string]interface{})
	for key, val := range i.hash {
		all[key] = val
	}

	return all
}
