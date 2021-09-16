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

	b.SetDefinition("counter.current", func(c di.Container) interface{} {
		return &Counter{
			count:     0,
			increment: c.GetParameter("counter.increment").(int),
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
	b.SetDefinition("counter.current #shared", func(c di.Container) interface{} {
		return &Counter{
			count:     0,
			increment: c.GetParameter("counter.increment").(int),
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
