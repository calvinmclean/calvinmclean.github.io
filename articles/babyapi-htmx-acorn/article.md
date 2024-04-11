## Introduction

**Note: Since writing this article, the Acorn SaaS product has been shutdown.**

In [a recent article](https://dev.to/calvinmclean/how-to-build-a-web-application-with-htmx-and-go-3183), I demonstrated how [`babyapi`](https://github.com/calvinmclean/babyapi), a library I created, makes it easy to write a TODO app with a RESTful API and HTMX frontend using only 150 lines of code. `babyapi` abstracts the HTTP handling based on a provided struct and serves HTMX templates for a dynamic frontend.

It works great in a tutorial, but you might have been left thinking: "what about persistent storage and running in the cloud?"

This simple tutorial will show you how to connect your `babyapi` application to Redis storage and quickly run in the cloud using the [free Sandbox](https://docs.acorn.io/sandbox) from [Acorn](https://www.acorn.io).

## Storage

The `babyapi.Storage` interface and the `SetStorage` modifier allow implementing any storage backend for your application. As mentioned in the previous article, the `babyapi/storage` package provides a generic implementation of the interface with helpers for setting up local file or Redis storage. This time, since the goal is to run in the cloud rather than just locally, we'll use the Redis version:

```go
db, err := storage.NewRedisDB(redis.Config{
    Server:   host + ":6379",
    Password: password,
})
if err != nil {
    return fmt.Errorf("error setting up redis storage: %w", err)
}

api.SetStorage(storage.NewClient[*TODO](db, "TODO"))
```

> The full example code for this tutorial is available in the [`babyapi` GitHub repository](https://github.com/calvinmclean/babyapi/blob/main/examples/todo-htmx/main.go)

With this simple addition, the TODO application is ready to connect to a Redis instance and run in the cloud.


## Acorn

If you're not already familiar with Acorn, I recommend checking out [the official docs](https://docs.acorn.io) to learn more about it! Basically, it is an app platform that makes it easy to deploy cloud applications and their dependencies by describing them in a simple Acornfile. Instead of configuring all of the required Kubernetes manifests to run our application in the cloud, we can just use an Acornfile.

First, we need a [Dockerfile](https://github.com/calvinmclean/babyapi/blob/main/examples/todo-htmx/Dockerfile) to create our app container. Then, instead of building and pushing this image to a container registry and writing a Kubernetes manifests (by copy/pasting from an online example if we're being honest), let's take a look at Acorn.

Before deploying the updated TODO app, we need the Redis database dependency. If you search "run redis in k8s", all of the top results look exhausting. Alternatively, the Acorn [documentation for Redis](https://docs.acorn.io/databases/redis) looks much simpler.

We can even use the Acornfile from the documentation's example with one little change to build and run the TODO app with Redis database:

```Acornfile
services: db: {
    image: "ghcr.io/acorn-io/redis:v7.#.#-#"
}

containers: app: {
    build: {
        context: "."
    }
    consumes: ["db"]
    ports: publish: "8080/http"
    env: {
        REDIS_HOST: "@{service.db.address}"
        REDIS_PASS: "@{service.db.secrets.admin.token}"
    }
}
```

Not only did this save us the effort of configuring everything for the TODO app, we even have the entire Redis dependency with a persistent volume and a random password. This is already way better than the no-volume, no-password K8s manifests I would thrown together for this example.


## Run it for real

Now that you have seen how it all works, you can run it for yourself using this button:

[![Run in Acorn](https://acorn.io/v1-ui/run/badge?image=ghcr.io+calvinmclean+babyapi-htmx-acorn&ref=calvinmclean&style=for-the-badge&color=brightgreen)](https://acorn.io/run/ghcr.io/calvinmclean/babyapi-htmx-acorn?ref=calvinmclean)

At the time of posting this, Acorn offers free Sandbox accounts to anyone with a GitHub account.

![Acorn UI](https://raw.githubusercontent.com/calvinmclean/calvinmclean.github.io/main/articles/babyapi-htmx-acorn/acorn_ui.png)

Once it has finished deploying, open the endpoint and append `/todos` in the URL to reveal the HTMX UI! You can even use the `babyapi` CLI from your terminal to create a new TODO and watch it show up in the UI automatically:

```shell
export ACORN_ADDR=http://COPY_ENDPOINT_FROM_ACORN

go run -mod=mod \
  github.com/calvinmclean/babyapi/examples/todo-htmx \
  -address $ACORN_ADDR \
  post TODOs '{"title": "use babyapi on Acorn!"}'
```


## Conclusion

The availability of app platforms like Acorn and easy-to-use libraries like [`babyapi`](https://github.com/calvinmclean/babyapi) take care of the boring and tedious parts of software engineering and let you focus on the things that interest you. Try them out for yourself and let me know what you have any feature requests or issues with `babyapi`.

Thanks for reading!
