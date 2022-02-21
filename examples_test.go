// Copyright (c) 2021 Santiago Garcia <sangarbe@gmail.com>.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package di_test

import (
	"fmt"
	"github.com/golossus/di"
)

type Counter struct {
	count     int
	Increment int `inject:"counter.increment"`
}

func (c *Counter) Incr() {
	c.count = c.count + c.Increment
}
func (c *Counter) Print() {
	fmt.Println(c.count)
}

func Example_basicUsage() {

	b := di.NewContainerBuilder()

	b.SetValue("counter.increment", 2)
	b.SetFactory("counter.current", func(c di.Container) interface{} {
		return &Counter{
			count:     0,
			Increment: c.Get("counter.increment").(int),
		}
	})

	c := b.GetContainer()

	//Get two instances of the service
	counter1 := c.Get("counter.current").(*Counter)
	counter1.Incr()
	counter1.Print()

	counter2 := c.Get("counter.current").(*Counter)
	counter2.Incr()
	counter2.Print()
	// Output:
	// 2
	// 2
}

func Example_sharedService() {

	b := di.NewContainerBuilder()

	b.SetValue("counter.increment", 2)

	//#shared is a reserved tag to declare singletons
	b.SetFactory("counter.current #shared", func(c di.Container) interface{} {
		return &Counter{
			count:     0,
			Increment: c.Get("counter.increment").(int),
		}
	})

	c := b.GetContainer()

	//Get the same instance twice
	counter1 := c.Get("counter.current").(*Counter)
	counter1.Incr()
	counter1.Print()

	counter2 := c.Get("counter.current").(*Counter)
	counter2.Incr()
	counter2.Print()
	// Output:
	// 2
	// 4
}

func Example_injectableStructs() {

	b := di.NewContainerBuilder()

	b.SetValue("counter.increment", 2)

	// Use the "inject" keyword to define the service to inject into a struct field label.
	// Service will be built from its Zero value, so like in this example, any field data of the passed struct will be
	// preserved. Here the value "1" for the field "count" won't be used, instead it will be "0".
	b.SetInjectable("counter.current", Counter{count: 1})

	c := b.GetContainer()

	//Get the same instance twice
	counter1 := c.Get("counter.current").(Counter)
	counter1.Incr()
	counter1.Print()
	counter1.Incr()
	counter1.Print()
	// Output:
	// 2
	// 4
}
