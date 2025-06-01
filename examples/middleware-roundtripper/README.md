# Middleware RoundTripper

**Most of this code is intended for educational purposes only. It is not production ready.**

The purpose of this example is to demonstrate a few common usecases for HTTP server middlewares and client-side round trippers.

It provides wrapper types to string together nested/sequential middlewares and roundtrippers. This makes it easy to show how different ordering impacts the effectiveness of these functions.

I encourage you to take a look at the tests which will demonstrate the different correct and incorrect uses of these.

Some incorrect uses aren't shown, like running a middleware that could panic before applying the RecoverMiddleware.

- If caching middleware is applied before sanitizing auth details, the auth details are cached
- If the logging or caching round tripper runs after applying auth to the request, the auth details are logged/cached
- If a server-side cache middleware is used before checking authentication, responses could be leaked to unauthorized requests
