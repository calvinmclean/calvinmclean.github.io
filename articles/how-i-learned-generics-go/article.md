## Introduction
Added in Go 1.18, generics were a controversial and long-awaited addition to the language. Go purists feared the change would encourage lazy practices and ruin Go's reputation for favoring simplicity over complexity. More recent Gophers were missing the utility of generics from other languages and eagerly awaited this update. I found myself somewhere in the middle.

I didn't have any desire to use generics initially, but understood that it could be useful in some cases. After working with Go long enough, I was already accustomed to thinking in a way that didn't involve generics and never thought twice about rewriting a `GetKeys(map[string]interface{})` function for the hundredth time. It wasn't until reading the release notes for Go 1.21, which added [`slices`](https://pkg.go.dev/slices) and [`maps`](https://pkg.go.dev/maps) packages to the standard library, that I started thinking more about generics. These packages are all built around generic type parameters and add common operations for the slice and map data structures. `maps.Keys` is currently in the [experimental version](https://pkg.go.dev/golang.org/x/exp/maps#Keys)!

Seeing how the standard library was able to use generics to expand its own functionality and alleviate some common pain points really opened my eyes to the potential of this feature. Originally, it seemed like generics were added to placate a few users, but these recent additions made me realize that generics are being taken seriously in Go and are worth a fresh look. I started thinking about how I could leverage generics in my own programs to reduce code duplication and make other improvements.

## First Attempt: API Layer
When thinking about duplicated code in my [`automated-garden`](https://github.com/calvinmclean/automated-garden) project, the first thing that comes to mind is all of my API handlers. The server side of this application implements a few straightforward CRUD APIs following RESTful principles. Each resource type implements handlers for the different HTTP verbs and mostly interacts with the storage layer. I created a very simple setup for the API handlers following this formula:
  1. Each resource has an ID, which is included in the URL
  2. Middleware is used to fetch a resource from storage and put it into the request context
  3. GET endpoints just use [`go-chi/render`](https://github.com/go-chi/render) to create the HTTP response
  4. Other endpoints, like PATCH and DELETE, read the resource from context and use it to perform additional actions

The only difference between the `GET` handlers is the type that is read from context and the response created. This, of course, sounds like a great use-case for generics! However, after thinking about it more and trying out a basic implementation, I realized it is possible to implement this using the `render.Renderer` interface that my types already implement:

```go
func getGardenFromContext(ctx context.Context) render.Renderer {
	return ctx.Value(gardenCtxKey).(*GardenResponse)
}

func get(getter func(context.Context) render.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resource := getter(r.Context())

		err := render.Render(w, r, resource)
		if err != nil {
			render.Render(w, r, ErrRender(err))
		}
	}
}

func routes(r *chi.Mux) {
    r.With(gardenContextMiddleware).Get("/", get(getGardenFromContext))
}
```

In order to implement this, I had to change the middleware to store the `GardenResponse` in the context instead of the raw `Garden` struct. The `getGardenFromContext` now has to return a `render.Renderer` so it can satisfy the input to `get` without additional type assertions. This makes the function less usable in other scenarios. One solution is replacing the `getter` input with just the context key, so the handler can use a more generalized `getRendererFromContext` function. Another option is to use a generic type parameter:

```go
func getGardenFromContext(ctx context.Context) *GardenResponse {
	// function body unchanged
}

func get[T render.Renderer](getter func(context.Context) T) http.HandlerFunc {
	// function body unchanged
}

func routes(r *chi.Mux) {
    r.With(gardenContextMiddleware).Get("/", get[*GardenResponse](getGardenFromContext))
}
```

The introduction of the generic type parameter adds a bit of clarity about the types being used and actually reduces the abstraction when compared to the interface-only approach. This is going to be appreciated by those concerned about unnecessary increases to complexity. Additionally, the `render.Renderer` interface is still used as a constraint for our generic type, so all of the benefits are still available. Finally, the `getGardenFromContext` function is returned to its original type so it can still be used outside of this response rendering scenario. In the end, it just comes down to personal preference.

The hidden truth for both implementations is that everything that isn't generalized is moved into the specific middlewares or the `Render` methods of each type. This is already how my application was designed, so it made the refactor pretty easy. However, this approach is not going to work as well for the more complicated handlers like `POST` and `PATCH` which do more than just returning the resource and have drastically different behaviors for each resource.

Ultimately, I decided to scrap this generic API idea, at least for now. While it slightly reduced code duplication, it doesn't work in all cases and furthers the separation between the application code and HTTP layer, which I see as an unnecessary abstraction. However, if I took the time to implement it as a generic API framework and rebuild my `automated-garden` API around it, that might be a different story.

## Next Attempt: Storage Layer
Determined to implement generics and delete some lines of code, I turned my sights to the storage layer of my application. I designed my storage around key-value pairs since I started by storing resources in YAML files. This eventually evolved to use [`madflojo/hord`](https://github.com/madflojo/hord) to interact with key-value data stores like Redis. The function to read a `*pkg.Garden` from storage looks like this:
```go
func (c *Client) getGarden(key string) (*pkg.Garden, error) {
	dataBytes, err := c.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("error getting Garden: %w", err)
	}

	var result pkg.Garden
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing Garden data: %w", err)
	}

	return &result, nil
}
```

This uses the `hord` client's `Get(key)` which returns a JSON `[]byte`. This is unmarshalled into the `pkg.Garden` struct. As you can probably imagine, the getters for other types are the same except the `var result pkg.Garden` line specifies a different type. Another potential usecase for generics! Or, once again, maybe it can be abstracted with interfaces.

In this case, the `var result pkg.Garden` line can be relocated outside of the function and passed in as a destination pointer. Instead of returning a `*pkg.Garden`, `get` will pass the destination pointer to `json.Unmarshal`. The function can now be used more generally:
```go
func (c *Client) GetGarden(id xid.ID) (*pkg.Garden, error) {
	var result *pkg.Garden
	err := c.get(gardenPrefix+id.String(), &result)
	return result, err
}

func (c *Client) get(key string, target interface{}) error {
	dataBytes, err := c.db.Get(key)
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	err = json.Unmarshal(dataBytes, &target)
	if err != nil {
		return fmt.Errorf("error parsing data: %w", err)
	}

	return nil
}
```

This is basically a wrapper around `json.Unmarshal` and using it will be familiar for any Go engineer. The same thing will work for `save` since it is similar to `json.Marshal`. Then, the only other function that touches the storage library is `getMultipleGardens` (and same for other types). Since I am using a simple key-value store instead of SQL tables, I have to get a list of keys and then fetch each one, which looks like this if I convert to use `interface{}`:
```go
func (c *Client) GetGardens() ([]*pkg.Garden, error) {
	var result []*pkg.Garden
	err := c.getMultiple(gardenPrefix, &result)
	return result, err
}

func (c *Client) getMultiple(prefix string, results []interface{}) error {
	keys, err := c.db.Keys()
	if err != nil {
		return fmt.Errorf("error getting keys: %w", err)
	}

	for _, key := range keys {
		if !strings.HasPrefix(key, gardenPrefix) {
			continue
		}

		var result interface{}
		err := c.get(key, &result)
		if err != nil {
			return fmt.Errorf("error getting Garden: %w", err)
		}
		if result == nil {
			continue
		}

        results = append(results, result)
	}

	return nil
}
```

This results in a compiler error: `cannot use result (variable of type []*pkg.Garden) as []interface{} value in argument to c.getMultiple`. This error occurs because `[]interface{}` does not behave in the same way as `interface{}`. Since it is a slice, each individual element needs to be typed separately. Even if this was valid, it would not work because the `c.get` call is using an emtpy `var result interface{}` and never knows which type to unmarshal into.

Now, this should be a case where generics can help since it is an abstraction on a data structure. Here is how it looks:
```go
func getOne[T any](c *Client, key string) (*T, error) {
	if c.db == nil {
		return nil, fmt.Errorf("error missing database connection")
	}

	dataBytes, err := c.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("error getting data: %w", err)
	}

	var result T
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing data: %w", err)
	}

	return &result, nil
}

func getMultiple[T any](c *Client, prefix string) ([]T, error) {
	keys, err := c.db.Keys()
	if err != nil {
		return nil, fmt.Errorf("error getting keys: %w", err)
	}

	results := []T{}
	for _, key := range keys {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		result, err := getOne[T](c, key)
		if err != nil {
			return nil, fmt.Errorf("error getting data: %w", err)
		}
		if result == nil {
			continue
		}

        results = append(results, *result)
	}

	return results, nil
}
```

Using generics here even simplifies the calling code:

```go
func (c *Client) GetGarden(id xid.ID) (*pkg.Garden, error) {
	return getOne[pkg.Garden](c, gardenKey(id))
}

func (c *Client) GetGardens(getEndDated bool) ([]*pkg.Garden, error) {
	return getMultiple[*pkg.Garden](c, getEndDated, gardenPrefix)
}
```

The result is a generic implementation that leaves the external interface of the storage client unchanged, but significantly reduces duplication internally. Take a look at the [full PR here](https://github.com/calvinmclean/automated-garden/pull/140). This is an ideal use of generics because it takes advantage of code reuse, but doesn’t expose the added complexity and abstraction to consumers of the package. Additionally, it was a very straightforward refactor because it strictly follows the advice of using generics ["where the only difference between the copies is that the code uses different types."](https://go.dev/blog/when-generics)

## Another Use-Case: Request Retries
This final example is the result of solving a problem that occurred organically rather than an attempt to brute-force generics into my existing application. While building a tool to [publish markdown files to DEV as articles using GitHub Actions](https://dev.to/calvinmclean/manage-dev-articles-with-git-and-github-actions-13md), I ran into rate limiting from the DEV API. Normally, HTTP client retries are done using a custom `http.Transport`, but this program uses a client library that doesn't allow setting the `Transport`. Instead, I turned to creating a wrapper function using generic type parameters for the response struct:
```go
type response interface {
	StatusCode() int
}

func doWithRetry[T response](do func() (T, error), numRetries int, initialWait time.Duration) (T, error) {
	for i := 1; i <= numRetries; i++ {
		result, err := do()
		if err != nil {
			return *new(T), err
		}

		if result.StatusCode() == http.StatusTooManyRequests {
			time.Sleep(initialWait * time.Duration(i))
			continue
		}

		return result, err
	}

	return *new(T), fmt.Errorf("exhausted retry limit %d", numRetries)
}

func (c *Client) getPublishedArticles() {
    resp, err := doWithRetry(func() (*api.GetUserPublishedArticlesResponse, error) {
		return c.GetUserPublishedArticlesWithResponse(context.Background(), nil)
	}, 5, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error getting articles: %w", err)
	}
}
```

You might notice that I don't include any type parameters when using the `doWithRetry` function. This is a nice feature of generics in Go that allows the type parameter to be inferred when it is used in multiple places. In this case, the return type of the `do` function has to be defined, so there is no need to duplicate this in the type parameter.

Another interesting aspect of this this example is that it takes advantage of Go's [implicit interfaces](https://dev.to/calvinmclean/the-magic-of-interfaces-in-go-5gga) to define an interface outside of the client library the defines all the structs and `StatusCode()` methods. The generic type parameter requires types that implement the interface so it can check response codes without additional type assertions. Using generics does not create an "either-or" scenario with interfaces, and often when using generics, interfaces are extremely useful to introduce constraints.

## Conclusion
I chose to learn about generics in Go by finding how the feature can improve my [`automated-garden`](https://github.com/calvinmclean/automated-garden) project. In this process, I learned how to use generics and more importantly, _when_ to use generics. I was reminded of the pitfalls that come from overusing abstraction and why I originally appreciated the language discouraging these practices, but I also see a lot of potential for this new feature.

Generic type parameters enable the simplification of small, tedious repetition and open the door to whole new levels of abstraction. In some cases, the feature provides a second way to achieve the same things that were already possible. In other cases, it unlocks whole new options that were previously unachievable. Most of the time, this level of abstraction is unnecessary, but it also allows for new and creative solutions. Although the potential to create overly-complex and abstracted code is increased, it is important for a programming language to provide these necessary tools. Ultimately, it is up to the programmer to decide how and when to use them. This reminds me of the famous quote: ["If all you have is a hammer, everything looks like a nail."](https://en.wiktionary.org/wiki/if_all_you_have_is_a_hammer,_everything_looks_like_a_nail) Luckily, with the addition of more tools to the language, we have much more than a hammer at our disposal.
