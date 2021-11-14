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
	increment int
}

func (c *Counter) Incr() {
	c.count = c.count + c.increment
}
func (c *Counter) Print() {
	fmt.Println(c.count)
}

func ExampleContainer_basicUsage() {

	b := di.NewContainerBuilder()

	b.SetParameter("counter.increment", 2)

	b.SetFactory("counter.current", func(c di.Container) interface{} {
		return &Counter{
			count:     0,
			increment: c.Get("counter.increment").(int),
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

func ExampleContainer_sharedService() {

	b := di.NewContainerBuilder()

	b.SetParameter("counter.increment", 2)

	//#shared is a reserved tag to declare singletons
	b.SetFactory("counter.current #shared", func(c di.Container) interface{} {
		return &Counter{
			count:     0,
			increment: c.Get("counter.increment").(int),
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
