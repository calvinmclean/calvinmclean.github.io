[Check out the full code referenced in this article on Github!](https://github.com/calvinmclean/calvinmclean.github.io/blob/main/examples/middleware-roundtripper)

Middleware and RoundTrippers are common web-development tools used in Go programs. Middleware operate on the server-side, while RoundTrippers are client-side. They are both reusable components that execute code before and/or after a request and response are processed.

In this article, I will provide some details about Middleware and RoundTrippers before providing interesting examples of each. Keep in mind that these are intended to be easily-understood examples and are not meant to be used in real applications as-is.

Then, tests will be used to demonstrate the correct and incorrect usage of these tools.

## Middleware

A middleware is just an HTTP handler (defined as `ServeHTTP(ResponseWriter, *Request)` [in Go's standard library](https://pkg.go.dev/net/http#Handler)) that calls another HTTP handler before and/or after doing some other actions. This is commonly used for logging and authentication. In the case of auth, it allows you to wrap every handler with the same auth middleware. If the middleware is applied to every handler in the application, the handlers can all operate assuming the request is already authenticated. A logging middleware allows you to observe every request's details as well as the response's status and duration. Go's standard library even provides a few middleware already: [StripPrefix](https://pkg.go.dev/net/http#StripPrefix) and [TimeoutHandler](https://pkg.go.dev/net/http#TimeoutHandler).

### Cache Middleware

This middleware caches responses on the server-side to prevent duplicate work. Implementing this simple middleware can result in a significant performance boost in applications with frequent `GET` requests. It can even reduce operational costs in cloud environments that charge for compute time, bandwidth, or traffic to databases. This can also be combined with an external cache, like Redis, to be functional across multiple instances.

```go
func CacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	cache := map[string][]byte{}

	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := fmt.Sprintf("%s_%s", r.Method, r.URL.String())

		// use cached response of the key matches
		if cached, ok := cache[cacheKey]; ok {
			serverLogger.Info("SERVER using cached response")
			w.Header().Add("X-Server-Cached", "true")
			w.Write(cached)
			w.WriteHeader(http.StatusOK)
			return
		}

		var buf bytes.Buffer
		// responseWriter is an implementation of the http.ResponseWriter
		// interface that can intercept the response body
		writer := &responseWriter{ResponseWriter: w, buffer: &buf}

		// Call the main HTTP handler
		next(writer, r)

		// Cache the response
		cache[cacheKey] = buf.Bytes()
	}
}
```

### Logger Middleware

This example simply logs request and response details.

```go
func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		serverLogger.Info(fmt.Sprintf("SERVER Request: %s %s", r.Method, r.URL.Path))

		var buf bytes.Buffer
		// responseWriter is an implementation of the http.ResponseWriter
		// interface that can intercept the response body
		writer := &responseWriter{ResponseWriter: w, buffer: &buf}

		// Call the main HTTP handler
		next(writer, r)

		serverLogger.Info(fmt.Sprintf("SERVER Response: %s %s %d %s", r.Method, r.URL.Path, writer.statusCode, time.Since(start)))
	}
}
```

### Auth Middleware

This is an simplified example of a middleware that reads the `Authorization` header and authenticates requests. It also removes the header so the secret/password is not at risk of logging or caching.

```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if auth != "password" {
			serverLogger.Warn("SERVER Authentication failed")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Forbidden")
			return
		}

		// remove header to avoid logging
		r.Header.Del("Authorization")
		serverLogger.Info("SERVER Authentication successful")

		// Call the main HTTP handler
		next(w, r)
	}
}
```

### Recovery Middleware

This middleware is essential in any webserver applications. Go runs a goroutine for each incoming request, but a `panic` will cause the whole application to crash. This middleware uses `recover()` to "catch" a panic, log it, and avoid crashing.

```go
func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				serverLogger.Error(fmt.Sprintf("SERVER panic: %v", err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		// Call the main HTTP handler
		next(w, r)
	}
}
```

### Sequential Middleware

This final example is a middleware that simplifies the composition of multiple middleware. Instead of using nested function calls, all of the middleware can be provided as a variadic argument.

```go
func SequentialMiddleware(middleware ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for _, mw := range slices.Backward(middleware) {
			final = mw(final)
		}
		return final
	}
}
```

Here is what it looks like to use it:
```go
handler := SequentialMiddleware(
	RecoveryMiddleware,
	AuthMiddleware,
	LoggerMiddleware,
	CacheMiddleware,
)(helloHandler)

// easier to read than:
RecoveryMiddleware(AuthMiddleware(LoggerMiddleware(CacheMiddleware(helloHandler))))
```

In addition to being easier to read and format, this can be used to create a slice of middleware (`[]func(http.HandlerFunc) http.HandlerFunc`) based on configuration values.


## RoundTrippers

[RoundTrippers](https://pkg.go.dev/net/http#RoundTripper) seem to be less common than middleware, but are very useful for client-side HTTP requests. It is called a RoundTripper because it handles the round-trip from the client to the server. This is shown by the interface's single method signature: `RoundTrip(*Request) (*Response, error)`. It takes a request and provides the response.

RoundTrippers are used by setting the `http.Client`'s `Transport` field. Unless you want to re-implement the code for establishing network connections between the client and the server, your RoundTripper implementations must, at some point, wrap Go's [DefaultTransport](https://pkg.go.dev/net/http#DefaultTransport). This, in my opinion, is a key distinction that makes RoundTrippers a bit more difficult to use than middleware. A middleware always uses your own code, but a custom RoundTripper needs to have the DefaultTransport somewhere in the chain. Therefore, almost any custom RoundTripper must wrap another RoundTripper.

### RoundTripperFunc

The `RoundTripperFunc` is similar to Go's `http.HandlerFunc` and simplfies the creation of RoundTrippers. Instead of creating a struct and implementing `RoundTrip`, this type allows implementing just the function.

```go
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (rt RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	if rt == nil {
		return http.DefaultTransport.RoundTrip(r)
	}
	return rt(r)
}
```

### Cache RoundTripper

This RoundTripper reduces network traffic and latency by using cached responses instead of making another round-trip to the server. Similar to the `CacheMiddleware`, this can provide a significant performance boost to applications with frequent external requests. When interfacing with vendor APIs that charge per-request, this can also signficantly reduce operational costs.

```go
func CacheRoundTripper(next http.RoundTripper) http.RoundTripper {
	cache := map[string]cachedResponse{}

	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		cacheKey := fmt.Sprintf("%s_%s", r.Method, r.URL.String())

		cachedResponse, ok := cache[cacheKey]
		if ok {
			clientLogger.Info("CLIENT using cached response")
			return cachedResponse.Response(), nil
		}

		// send the actual request
		resp, err := next.RoundTrip(r)

		// cache the response
		cache[cacheKey] = newCachedResponse(resp)
		return resp, err
	})
}
```

The `cachedResponse` type is excluded here to keep the example short. This is just a new struct that copies the `http.Response` and its body. You can check out the full example code [here](https://github.com/calvinmclean/calvinmclean.github.io/blob/main/examples/middleware-roundtripper).

### Log RoundTripper

This is very similar to the log middleware and just logs the request and response details.

```go
func LogRoundTripper(next http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		clientLogger.Info(fmt.Sprintf("CLIENT Request: %s %s %v", r.Method, r.URL, r.Header))
		start := time.Now()

		// Send the actual request
		resp, err := next.RoundTrip(r)

		duration := time.Since(start)
		if err != nil {
			clientLogger.Error(fmt.Sprintf("CLIENT Request failed after %s: %v", duration, err))
			return nil, err
		}
		clientLogger.Info(fmt.Sprintf("CLIENT Response: %s %s %d in %s", r.Method, r.URL, resp.StatusCode, duration))
		return resp, nil
	})
}
```

### Auth RoundTripper

This is a simplified example of using a RoundTripper for adding auth to a request. In this case, it just adds a password to the `Authorization` header. Even though this is a simple case, it is still useful because only one part of your application needs to be aware of the password/secret and the rest will reuse the RoundTripper.

```go
func AuthRoundTripper(next http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		r.Header.Add("Authorization", "password")
		return next.RoundTrip(r)
	})
}
```

In a real-world scenario, an auth RoundTripper can handle more complex tasks like using a refresh token to get an updated access token when it expires.

Go's [documentation on RoundTripper says](https://pkg.go.dev/net/http#RoundTripper):
> RoundTrip should not attempt to handle higher-level protocol details such as redirects, authentication, or cookies.
> RoundTrip should not modify the request, except for consuming and closing the Request's Body

However, we'll ignore that here since it is overly-cautious and severely limits the RoundTripper's functionality. This is likely a constraint because the `*http.Request` is a pointer and modifying it may have unexpected consequences. In production-ready RoundTrippers, `r.Clone(r.Context())` can be used to create a new `*http.Request` that is safe to modify.


### Sequential RoundTripper

Similar to the `SequentialMiddleware`, this is a helper that simplifies the composition of multiple RoundTrippers. It is especially useful with the `RoundTripperFunc`. It also ensures that the last step in the RoundTripping is `http.DefaultTransport`.

```go
func SequentialRoundTripper(rts ...func(http.RoundTripper) http.RoundTripper) http.RoundTripper {
	result := http.DefaultTransport

	for _, rt := range slices.Backward(rts) {
		result = RoundTripperFunc(rt(result).RoundTrip)
	}

	return result
}
```

Now, we can use a few RoundTrippers like this:
```go
SequentialRoundTripper(
	LogRoundTripper,
	CacheRoundTripper,
	AuthRoundTripper,
)

// easier to read than:
LogRoundTripper(CacheRoundTripper(AuthRoundTripper(http.DefaultTransport)))
```

In addition to being easier to read and format, this can be used to create a slice of RoundTrippers based on configuration values.


## Tests

Now we have a toolkit of middleware and RoundTrippers that can be used for an HTTP server and a client. In order to demonstrate this, we can write some Go tests.

### Correct Setup

```go
func TestGood(t *testing.T) {
	handler := SequentialMiddleware(
		RecoveryMiddleware,
		AuthMiddleware,
		LoggerMiddleware,
		CacheMiddleware,
	)(helloHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("SuccessfulResponse", func(t *testing.T) {
		client := NewClientWithRoundTrippers(
			LogRoundTripper,
			CacheRoundTripper,
			AuthRoundTripper,
		)

		body, resp, err := doTestRequest(client, server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectations{
			StatusCode: http.StatusOK,
			Body:       "Hello, World!\n",
		}.assert(t, body, resp)

		// request again to use client cache
		body, resp, err = doTestRequest(client, server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectations{
			StatusCode:      http.StatusOK,
			Body:            "Hello, World!\n",
			ClientCacheUsed: true,
		}.assert(t, body, resp)
	})

	t.Run("DisableClientCacheToDemoServerCache", func(t *testing.T) {
		client := NewClientWithRoundTrippers(
			LogRoundTripper,
			AuthRoundTripper,
		)

		body, resp, err := doTestRequest(client, server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectations{
			StatusCode:      http.StatusOK,
			Body:            "Hello, World!\n",
			ServerCacheUsed: true,
		}.assert(t, body, resp)
	})

	t.Run("ForbiddenWithoutAuthRoundTripper", func(t *testing.T) {
		client := NewClientWithRoundTrippers(
			LogRoundTripper,
		)

		body, resp, err := doTestRequest(client, server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectations{
			StatusCode: http.StatusForbidden,
			Body:       "Forbidden\n",
		}.assert(t, body, resp)
	})
}
```

This test starts by setting up the HTTP server with the middleware in the correct order:
1. `RecoveryMiddleware`: this is the first code to run. It defers `recover()` to catch any `panic` that happens down the line
2. `LoggerMiddleware`: logs the request and response details. Running this second ensures that every request and response is logged, even ones that fail auth
3. `AuthMiddleware`: before doing anything else, the request should be authenticated
4. `CacheMiddleware`: finally, cache the response. Any requests that use a cached response will still be authenticated and logged

Next, a few different subtests are used to show different client scenarios. These all use the same general ordering, but exclude the cache or auth RoundTrippers to demonstrate different outcomes:
1. `LogRoundTripper`: always logs request/response
2. `CacheRoundTripper`: cache the response. This is removed to demonstrate that the server-side cache also works
3. `AuthRoundTripper`: adds the `Authorization` header to the request. It is removed to show that the request fails without it

The output of this test shows that the client and server both log the request and response.


### Incorrect Setup

```go
func TestBad_ServerCacheExposesAccess(t *testing.T) {
	// The server will cache responses before checking authentication. After caching a response, the next request
	// will receive this response even if it does not authenticate successfully
	handler := SequentialMiddleware(
		CacheMiddleware,
		LoggerMiddleware,
		AuthMiddleware,
		RecoveryMiddleware,
	)(helloHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("PopulateCacheWithAuth", func(t *testing.T) {
		client := NewClientWithRoundTrippers(
			LogRoundTripper,
			AuthRoundTripper,
		)

		body, resp, err := doTestRequest(client, server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectations{
			StatusCode: http.StatusOK,
			Body:       "Hello, World!\n",
		}.assert(t, body, resp)
	})

	// Server cache is used before AuthMiddleware, allowing unrestricted access
	t.Run("RequestWithoutAuth", func(t *testing.T) {
		client := NewClientWithRoundTrippers(
			LogRoundTripper,
		)

		body, resp, err := doTestRequest(client, server.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectations{
			StatusCode:      http.StatusOK,
			Body:            "Hello, World!\n",
			ServerCacheUsed: true,
		}.assert(t, body, resp)
	})
}
```

In this test, the middleware are initialized in reverse order. This has a few issues:
1. While the `RecoveryMiddleware` will still handle panics in the main handler, it will not handle panics in any of the other middleware
2. Since the `CacheMiddleware` is the first to run, it will immediately respond with a cached response before checking auth. This means that unauthorized request will be able to access the server
3. Additionally, since the `LoggerMiddleware` runs after the cache, we won't even know that these unauthorized requests are occurring

This is a trivial example, but we can see how a simple mistake like incorrect middleware order can have catatrophic results. Similar, but less-impactful errors can occur with the RoundTrippers. If we use the `AuthRoundTripper` before the `CacheRoundTripper`, our authentication credentials will be saved in the cache, potentially leaking sensitive details. Also, we will lose a benefit of the cache if the `AuthRoundTripper` is refreshing access tokens.


## Conclusion

Middleware and RoundTrippers are simple and common tools in Go. While the topic is familiar for many seasoned Gophers, it's beneficial to revisit the basics and remember to carefully consider how they are used. These are just a few of the usecases and possible pitfalls of using middleware and RoundTrippers. Here are a few more possible ways to leverage them:
- Automatic retries for 500-level errors (RoundTripper). Make sure to use this after the `CacheRoundTripper` so you can use the cache before encountering and retrying any errors
- OpenTelemetry tracing and metrics (RoundTripper and middleware)
- Throttling and rate-limiting. Use with a middleware to protect your server from abuse, or implement in a RoundTripper to avoid overloading another server. Similar to the retries, this RoundTripper should run after the `CacheRoundTripper` since the cache will already reduce the number of requests
- Client-side load balancing by choosing from a pool of upstream URLs
- Testing. The [`go-vcr`](https://github.com/dnaeon/go-vcr/tree/v4) library leverages a `RoundTripper` to record interactions with external servers and then replay them using `httptest.Server`. Make sure to run this RoundTripper before the `AuthRoundTripper` so your `Authorization` header isn't saved in a test fixture! Different tokens would also cause requests to not match

Let me know some of your favorite or novel usecases for middleware and RoundTrippers in Go!

[Check out the full code referenced in this article on Github!](https://github.com/calvinmclean/calvinmclean.github.io/blob/main/examples/middleware-roundtripper)
