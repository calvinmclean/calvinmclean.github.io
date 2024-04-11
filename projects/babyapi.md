# BabyAPI

[![babyapi](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/calvinmclean/babyapi)

`babyapi` is a super simple framework that automatically creates an HTTP API for create, read, update, and delete operations on a struct. It is intended to make it as easy as possible to go from nothing to a fully functioning REST API, but also aims to allow enough flexibility for creating real web applications. It also has built-in integrations for using Redis as a key-value store so you can quickly and easily create an HTTP API with persistent storage.

I got tired of re-writing boilerplate code for HTTP handling in my `automated-garden` project. I also was interested in learning about generics in Go, so I extracted some of the generic behaviors into reusable methods. This eventually evolved into `babyapi`.

You can read more about `babyapi` in [this article that I wrote](https://dev.to/calvinmclean/the-easiest-way-to-create-a-rest-api-with-go-20bo) and the [GitHub repository](https://github.com/calvinmclean/babyapi).
