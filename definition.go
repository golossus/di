// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di

//definition represents a service factory definition with additional metadata.
type definition struct {
	Factory         func(Container) interface{}
	Tags            *itemHash
	Shared, Private bool
	Priority        int16
	AliasOf         *definition
}
