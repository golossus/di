# Golossus DI

![ci workflow](https://github.com/golossus/di/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/golossus/di/branch/master/graph/badge.svg?token=DGXZAW5PZF)](https://codecov.io/gh/golossus/di)
[![Go Report Card](https://goreportcard.com/badge/github.com/golossus/di)](https://goreportcard.com/report/github.com/golossus/di)
[![CodeFactor](https://www.codefactor.io/repository/github/golossus/di/badge)](https://www.codefactor.io/repository/github/golossus/di)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=golossus_di&metric=alert_status)](https://sonarcloud.io/dashboard?id=golossus_di)

Dependency Injection container for go projects (wip).

<p align="center">
    <a href="https://www.golossus.com" target="_blank">
        <img height="100" src="https://avatars2.githubusercontent.com/u/58183018">
    </a>
</p>

[Golossus][1] is a set of reusable **Go modules** to facilitate the creation of applications, specially suited for web,
leveraging Go's standard packages.

The Dependency Injection (DI) module provides a seamless mechanism to declare a dependency injection service container
(or IoC container) for go projects. DI containers significantly help to manage dependencies within the application, so
it's considered a convenient pattern to increase code maintainability.

DI containers handle the recursive creation (or complete lifecycle) of "objects" (usually called services) and their
required dependencies in an automatic fashion. Golossus DI main features are:

* Reflection based container.
* Declare values, factories, aliases or injectable structs.
* Associate tags to services and retrieve them by priority.
* Simple and minimalistic API.
* Supports singletons.
* Defines Provider & Resolver interfaces to allow better organization of definitions.
* Control of services access using private or public scope.
* More to come...

Installation
------------

The DI package is just a common Go module, it can be installed as any other Go module, go to your project root folder on
your preferred shell and type:

```shell
go get github.com/golossus/di
```

To get more information just review the official [Go blog][2] regarding this topic.

Usage
-----

In order to create a service container, first thing to is to create a new instance of a container builder exposed by the
DI module. After all the services have been defined, the resolved container can be requested from the builder. Once the
container instance is ready, services can be retrieved form it, mainly by key.

> It is required to type cast the resultant service because the container is a generic like factory and only retrieves
> `interface{}` types. If the type conversion is incompatible the code will panic. This is in addition to performance
> the main flaws of using reflection for this kind of pattern.

```go
package main

func main() {
	// create a builder instance
	builder := di.NewContainerBuilder()

	// configure service definitions
	...

	// resolved and get the final container
	container := builder.GetContainer()

	// get a service, this might panic if type conversion fails
	service := container.Get("my.service.key").(ServiceType)

	...
}
```

The builder exposes a simple API to configure the creation of any desired object. Golossus DI builder basically
considers the definition of four types of objects because of their differentiated nature: values, factories, aliases and
injectable structs.

### Setting Values

Values are just instances of pre-built objects. Sometimes it's handy to add any previously created object (it might be
anything) into the container instead of wiring the creation of it by means of the builder API. For example, this might
be helpful for primitive values that are required as dependencies of other services or, to add an object retrieved from
another package.

Values can be anything (scalars, structs, functions, ...) as far as they are created beforehand. Use the `SetValue`
method of the builder to bind the value to some key:

```go
package main

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	...
	container := builder.GetContainer()
	...
	fromEmail := container.Get("email.from").(string)
}
```

> In the example above, the container will retrieve a copy of the `from` value each time it's requested using `container.Get("email.from")`.
> The reason is simple, function return values are passed by value the same way function arguments do in Go. If you want
> the same instance of the value to be returned on each retrieval, refer to the shared services section. Tip: use a pointer.

### Setting Factories

Factories are the de-facto approach to configure object creation. Factories are simply functions that receive
a `Container`
instance implementation as their only argument and return a single `interface{}`. Inside factories, the container can be
used to retrieve any dependency required by the returned object from the container.

Factories will be called during object creation and not during object configuration. This means, you can require any
service from the container by key even if the key doesn't exist yet, as far as on creation time the service is available
on the container (code will panic if not the case). In order to declare factories for the container use the method
`SetFactory` of the builder instance.

> In order to avoid lack of service definition and unexpected side effects on run time, a special method `MustBuild` has
> been provided to the container. Refer to his section to know more about this method.

```go
package main

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	builder.SetFactory("email.mailer", func(c Container) interface{} {
		return Mailer{from: c.Get("from").(string)}
	})
	...
	container := builder.GetContainer()
	mailer := container.Get("email.mailer").(Mailer)
}
```

> In the example above, the container will retrieve a copy of the `from` value each time it's requested using `container.Get("email.from")`.
> The reason is simple, function return values are passed by value the same way function arguments do in Go. If you want
> the same instance of the value to be returned on each retrieval, refer to the shared services section. Tip: use a pointer.

### Setting Injectable structs

Injectable structs are common structs whose fields are labeled with a special label `inject`. These are handy to be used
when the object to retrieve from the container is a struct and a more declarative way to define their dependencies is
preferred. Injectable structs are automatically injected with the services represented by the keys used after
the `inject`
label. Use the method `SetInjectable` of the builder instance to pass an injectable struct instance with at least one
eported field labeled to be injected.

> Only exported fields of the structs can be injected, and the final struct will be created from its initial "zero" value.
> Trying to inject an unexported field will panic, same way if types of the dependencies and fields are not equivalent, or
> no field is labeled. Additionally, no matter the values of a struct instance has when passed to the `SetInjectable`,
> they won't be used in the retrieved struct from the container.

```go
package main

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	builder.SetInjectable("email.mailer", &Mailer{})
	...
	container := builder.GetContainer()
	mailer := container.Get("email.mailer").(*Mailer)
}
```

> Notice that we can use pointers to services as well, as far as type conversion is done properly there's no limitation
> on the kind of objects the container can handle.

### Setting Aliases

Aliases are a simple way to define another service based on an existing definition. Aliases can have different tags (
refer to tags section) or scope (refer to private services section) than the aliased service. Just use the `SetAlias`
method of the builder instance to define a new alias of an existing service definition.

> Trying to define an alias of a not existing definition will panic. Trying to define an alias with a key already used
> by another service definition will panic as well.

> Trying to define an alias with a key already used by another alias will replace the former alias. Same way, services
> will always replace aliases if using same keys.

```go
package main

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	builder.SetInjectable("email.mailer", &Mailer{})
	builder.SetAlias("mailer.default", "email.mailer")
	...
	container := builder.GetContainer()
	mailer := container.Get("mailer.default").(*Mailer)
}
```

### Adding tags to services

Tags can be added to service's definition as a form of metadata. There are two ways to associate tags to services: as
part of the service key, or as setter method argument.

Tags can be included as **part of a service's key** by preceding them with a `#` charracter. Additionally, specific
values for tags can also be included preceding them with the `=` character. The actual key of a service will be the
prefix part of a key until the first `#`. Subsequently, the tag name will be the following prefix of the key until the
first `=` or `#` found. Either the key portion, tag name or tag value will be blank space trimmed.

Tags can also be included as a map argument on the setters methods of the builder. Tags and values added through this
method will take precedence over the ones set through the key.

Services can be retrieved from the container by tag name and value using the method `GetTaggedBy`. The container will
return a list of services in this case

```go
package main

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	builder.SetInjectable("email.mailer.1 #score=10 #mailing", &Mailer{}) // <- tags on key
	builder.SetInjectable("email.mailer.2", &Mailer{}, map[string]string{ // <- tags on argument
		"mailing": "",
		"score":   "1",
	})
	...
	container := builder.GetContainer()

	mailers := container.GetTaggedBy("mailing")
	mailer1 := mailers[0].(*Mailer)
	mailer2 := mailers[1].(*Mailer)
	
	// or 
	mailers := container.GetTaggedBy("score", "10")
	mailer := mailers[0].(*Mailer)

}
```

Tags are also important because a reserved set of tags can be used to configure the behaviour of the service definitions.
Go to following sections to know more about them.

### Shared services (singletons)

By default, service definitions are configured as **not shared** services. This means, each time a service is requested 
from the container, the container will create a new instance of that service. If the same instance of a service must be
reused once built, then the service definition must be set as **shared**. This can be done by using the reserved tag name
`shared` or its exported const equivalent `TagShared`. Adding this tag to a service definition will make the service to
build only once a service instance and reuse it on consecutive retrievals.

This is the approach to define and retrieve **singletons** from the container. The only detail to bare in mind is that, 
if a singleton is required, the service should be a pointer. If we don't use a pointer for a shared service, inevitably
we will get a copy of the instance as a result value of the corresponding function. This can have unexpected results if
not aware.

```go
package main

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	builder.SetInjectable("email.mailer.singleton1 #shared", &Mailer{})             // <- tag shared on key
	builder.SetInjectable("email.mailer.singleton2 #shared=true", &Mailer{})        // <- tag with boolean value
	builder.SetInjectable("email.mailer.not.singleton1 #shared", Mailer{})          // <- not a pointer, so not singleton
	builder.SetInjectable("email.mailer.not.singleton2 #shared=false", Mailer{})    // <- not shared, so not singleton
	...
}
```

### Private services (scope)

By default, service definitions are **public**. This means, a service can be retrieved directly from the container. This
might not always be the desire behaviour, because we might need to restrict the access to some services that are only
relevant to build other services. 

In order to make a service **private** a reserved tag `private` can be used or its corresponding exported const `TagPrivate`.
Private services can not be retrieved from the built container neither using `Get` nor `GetTaggedBy` methods, but they 
can be retrieved with those same methods on service definitions.

> Remember that aliases have their own set of tags so you can have a **private** service and a **public** alias of the
> same service. This might be useful to expose a small set of well know public services by means of predefined aliases.

```go
package main

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetValue("email.from", from)
	builder.SetInjectable("email.mailer.singleton1 #shared", &Mailer{})             // <- tag shared on key
	builder.SetInjectable("email.mailer.singleton2 #shared=true", &Mailer{})        // <- tag with boolean value
	builder.SetInjectable("email.mailer.not.singleton1 #shared", Mailer{})          // <- not a pointer, so not singleton
	builder.SetInjectable("email.mailer.not.singleton2 #shared=false", Mailer{})    // <- not shared, so not singleton
	...
}
```

### Setting All at once

A convenient method `SetAll` of the builder can be used to set all types of bindings in a single call to facilitate code
organization. The only requirement is that a special tag identifying the type of definition must be used (if not
provided `factory` type will be used by default).

> Constants `TagValue`, `TagFactory`, `TagAlias` and `TagInject` can be used instead of their counterpart string values 
> `value`, `factory`, `alias` and `inject`.

```go
package main

import "github.com/golossus/di"

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetAll([]di.Binding{
		{Key: "email.from #value", Target: from},                                   // <- tag as key
		{Key: "email.mailer #inject", Target: &Mailer{}},
		{Key: "mailer.default", Target: "email.mailer", Tags: map[string]string{    // <- tag as metadata
			di.TagAlias: "",                                                        // <- tag constant
		}},
	}...)

	container := builder.GetContainer()
	mailer := container.Get("mailer.default").(*Mailer)
}
```

### Container check with MustBuild

As mentioned before, due to the nature of reflection in Go, we can have panics while building our services through the
container or while doing invalid type conversions of the retrieved services. In order to minimiz surprises on run time, 
a special method of the container called `MustBuild` is provided.

This method can be used to build all the services defined in the container in a row. If no panic occurs during the 
construction of each service we can be sure the container is safe. At least, we can ensure the container won't panic while
building any of its services. 

This method can be used at the application startup, so that if it panics, we don't have a running application which
eventually will fail because of a bad service definition.

```go
package main

import "github.com/golossus/di"

type Mailer struct {
	from string `inject:"email.from"`
}

func main() {
	from := "from@email.com"

	builder := di.NewContainerBuilder()
	builder.SetAll([]di.Binding{
		{Key: "email.from #value", Target: from},                                   
		{Key: "email.mailer #inject", Target: &Mailer{}},
		{Key: "mailer.default", Target: "email.mailer", Tags: map[string]string{    
			di.TagAlias: "",                                                        
		}},
	}...)

	container := builder.GetContainer()
	container.MustBuild(true)               // dry run set to true, will fail if definition error                                               
	
}
```

Community
---------

* Join our [Slack][5] to meet the community and get support.
* Follow us on [GitHub][6].
* Read our [Code of Conduct][7].

Contributing
------------

Golossus is an Open Source project. The Golossus team wants to enable it to be community-driven and open
to [contributors][8]. Take a look at [contributing documentation][9].

Security Issues
---------------

If you discover a security vulnerability within any Golossus module, please follow our[disclosure procedure][10].

About Us
--------

DI module development is led by the Golossus Team [Leaders][12] and supported by [contributors][8].

[1]: https://github.com/golossus

[2]: https://blog.golang.org/using-go-modules

[5]: https://join.slack.com/t/golossus/shared_invite/zt-db4brnes-M8q1Lw2ouFT5X~gQg69NQQ

[6]: https://github.com/golossus

[7]: ./CODE_OF_CONDUCT.md

[8]: ./CONTRIBUTORS.md

[9]: ./CONTRIBUTING.md

[10]: ./CONTRIBUTING.md#reporting-a-security-issue

[12]: ./CONTRIBUTING.md#leaders