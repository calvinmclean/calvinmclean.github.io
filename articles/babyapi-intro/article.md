## Introduction

Go is known for being easy to learn and providing the fastest path to production. With HTTP functionality built-in to the standard library, you have everything you need without any external dependencies. The lower-level access to HTTP can be a breath of fresh air if you are accustomed to bloated frameworks, but it results in a lot of boilerplate code which eventually becomes tiring to write over and over. This is where [`babyapi`](https://github.com/calvinmclean/babyapi) comes in. I created this library to enable super simple creation of REST APIs without forcing a specific structure on your whole application. It's so easy that a baby could do it!


## Baby API

All you need to get started with `babyapi` is a struct that implements the `babyapi.Resource` interface. This is easily achieved by extending the `babyapi.DefaultResource` struct. After completing this simple step, you will have an HTTP API capable of create, read, update, and delete operations. Additionally, `babyapi` provides an HTTP client to interact with the API, a CLI to easily use this client, and some shortcuts for unit testing.

`babyapi` can also take you beyond the basics. It provides plenty of options for extending the API behavior, adding validation, and even custom API routes. Any storage backend can be integrated by implementing the simple `Storage` interface.

This article just scratches the surface of what `babyapi` is capable of. If you are interested in learning more, keep an eye out for my future articles and star the [GitHub repository](https://github.com/calvinmclean/babyapi).


## Getting Started

The goal of `babyapi` is to be so easy that a baby could do it. As previously mentioned, `babyapi.DefaultResource` already implements the required interface, so it can be used as a starting point for simple resource types. Besides simply implementing the interface, this default struct implements some validations around the ID and uses [`rs/xid`](https://github.com/rs/xid) to create a unique identifer on new resources.

Here is a simple example that extends the default to create an API for TODO items:

```go
package main

import "github.com/calvinmclean/babyapi"

type TODO struct {
    babyapi.DefaultResource
    Title       string
    Description string
    Completed   bool
}

func main() {
    api := babyapi.NewAPI[*TODO](
        "TODOs", "/todos",
        func() *TODO { return &TODO{} },
    )
    api.RunCLI()
}
```

Next, all you need to do is run the server and use the CLI to interact with it:

```shell
go run main.go serve
```

<img alt="Simple Example" src="https://raw.githubusercontent.com/calvinmclean/babyapi/main/examples/simple/simple.gif" width="600" />

There are too many features to cover in this article, so the best ways to learn more are:
  - [Getting Started tutorial](https://github.com/calvinmclean/babyapi#getting-started)
  - [Additional examples](https://github.com/calvinmclean/babyapi#examples)
  - [`pkg.go.dev` documentation](https://pkg.go.dev/github.com/calvinmclean/babyapi)
  - My future articles :)


## Client

The struct used to create the server also enables a built-in client for interacting with the `babyapi` server. The client can also be used in the application or other Go applications interacting with your API. It supports top-level and nested API resources and allows easy access to all CRUD endpoints.

```go
// Create a client from an existing API struct (mostly useful for unit testing):
client := api.Client(serverURL)

// Create a client from the Resource type:
client := babyapi.NewClient[*TODO](addr, "/todos")
```

```go
// Create a new TODO item
todo, err := client.Post(context.Background(), &TODO{Title: "use babyapi!"})

// Get an existing TODO item by ID
todo, err := client.Get(context.Background(), todo.GetID())

// Get all incomplete TODO items
incompleteTODOs, err := client.GetAll(context.Background(), url.Values{
    "completed": []string{"false"},
})

// Delete a TODO item
err := client.Delete(context.Background(), todo.GetID())
```
 
The client provides more general methods for interacting with server: `MakeRequest` and `MakeRequestWithResponse`. The client is also cutomizable. You can replace the underlying `http.Client` to use a custom `RoundTripper` or use `SetRequestEditor` to add a pre-request function that modifies the `http.Request`.


## Storage

You can bring any storage backend to `babyapi` by implementing the `Storage` interface. By default, the API will use the built-in `MapStorage` which just uses an in-memory map.

In an effort to provide actual persistent storage out of the box, the `babyapi/storage` package uses [`madflojo/hord`](https://github.com/madflojo/hord) to support a variety of key-value store backends. Additionally, `babyapi/storage` provides helper functions for initializing the `hord` client for Redis or file-based storage.

```go
db, err := storage.NewFileDB(hashmap.Config{
    Filename: "storage.json",
})
db, err := storage.NewRedisDB(redis.Config{
    Server: "localhost:6379",
})

api.SetStorage(storage.NewClient[*TODO](db, "TODO"))
```


## Conclusion & Roadmap

`babyapi` is a very new project and still has a lot of potential for new features and improvements. However, my goal is to keep it as simple and barebones as possible, so I plan to focus on usability and testing improvements. I will add new features if use-cases or feature requests come up. These are some of the things coming next:
  - Improve CLI to have better usage instructions and implement flags when multiple IDs are required for nested APIs
  - Improve default logging and implement log-levels
  - Allow CLI to access non-CRUD endpoints (endpoints added by `AddCustomRoute` or `AddCustomIDRoute`)
  - Create more automated testing. Ideally, I will create a generator that reads your source file using `babyapi` and generates a test structure with some default tests and room to extend them

Also keep an eye out for more articles that dig deeper into some of the core features of `babyapi`!
