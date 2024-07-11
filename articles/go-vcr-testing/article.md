## Introduction

As a software engineer, you are probably familiar with writing code to interact with external HTTP services. After all, it is one of the most common things we do! Whether it's fetching data, processing payments with a provider, or automating social media posts, our applications almost always involve external HTTP requests. In order for our software to be reliable and maintainable, we need a way to test the code responsible for executing these requests and handling the errors that could occur. This leaves us with a few options:
  - Implement a client wrapper that can be mocked by the main application code, which still leaves a gap in testing
  - Test response parsing and handling separate from actual request execution. While it's probably a good idea to test this lower-level unit individually, it'd be nice if that could easily be covered along with the actual requests
  - Move tests to integration testing which can slow down development and is unable to test some error scenarios and may be impacted by the reliability of other services

These options aren't terrible, especially if they can all be used together, but we have a better option: VCR testing.

VCR testing, named after the videocassette recorder, is a type of mock testing that generates [test fixtures](https://en.wikipedia.org/wiki/Test_fixture) from actual requests. The fixtures record the request and response to automatically reuse in future tests. Although you might have to modify the fixtures afterwards to handle dynamic time-based inputs or remove credentials, it is much simpler than creating mocks from scratch. There are a few additional benefits to VCR testing:
  - Execute your code all the way down to the HTTP level, so you can test your application end-to-end
  - You can take real-world responses and modify the generated fixtures to increase response time, cause rate limiting, etc. to test error scenarios that don't often occur organically
  - If your code uses an external package/library for interacting with an API, you might not know exactly what a request and response look like, so VCR testing can automatically figure that out
  - Generated fixtures can also be used for debugging tests and making sure your code executes the expected request


## Deeper Dive using Go

Now that you see the motivation behind VCR testing, let's dig deeper into how to implement it in Go using [`dnaeon/go-vcr`](https://github.com/dnaeon/go-vcr).

This library integrates seamlessly into any HTTP client code. If your client library code doesn't already allow setting the `*http.Client` or the Client's `http.Transport`, you should add that now.

For those that aren't familiar, an `http.Transport` is an implementation of `http.RoundTripper`, which is basically a client-side middleware that can access the request/response. It is useful for implementing automatic retries on 500-level or 429 (rate-limit) responses, or adding metrics and logging around requests. In this case, it allows `go-vcr` to re-reoute requests to its own in-process HTTP server.

### URL Shortener Example

Let's get started on a simple example. We want to create a package that makes requests to the free https://cleanuri.com API. This package will provide one function: `Shorten(string) (string, error)`

Since this is a free API, maybe we can just test it by making requests directly to the server? This might work, but can result in a few problems:
  - The server has a rate limit of 2 requests/second which could be an issue if we have a lot of tests
  - If the server goes down or takes awhile to respond, our tests could fail
  - Although the shortened URLs are cached, we have no guarantee that we will get the same output every time
  - It's just rude to send unnecessary traffic to a free API!

Ok, what if we create an interface and mock it? Our package is incredibly simple, so this would overcomplicate it. Since the lowest-level thing we use is `*http.Client`, we would have to define a new interface around it and implement a mock.

Another option is to override the target URL to use a local port served by `httptest.Server`. This is basically a simplified version of what `go-vcr` does and would be sufficient in our simple case, but won't be maintainable in more complex scenarios. Even in this example, you'll see how managing generated fixtures is easier than managing different mock server implementations.

Since our interface is already defined and we know some valid input/output from trying the UI at https://cleanuri.com, this is a great opportunity to practice [test-driven development](https://dev.to/calvinmclean/test-driven-api-development-in-go-1fb8). We'll start by implementing a simple test for our `Shorten` function:

```go
package shortener_test

func TestShorten(t *testing.T) {
	shortened, err := shortener.Shorten("https://dev.to/calvinmclean")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if shortened != "https://cleanuri.com/7nPmQk" {
		t.Errorf("unexpected result: %v", shortened)
	}
}
```

Pretty easy! We know that the test will fail to compile because `shortener.Shorten` is not defined, but we run it anyways so fixing it will be more satisfying.

Finally, let's go ahead and implement this function:

```go
package shortener

var DefaultClient = http.DefaultClient

const address = "https://cleanuri.com/api/v1/shorten"

// Shorten will returned the shortened URL
func Shorten(targetURL string) (string, error) {
	resp, err := DefaultClient.PostForm(
		address,
		url.Values{"url": []string{targetURL}},
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	var respData struct {
		ResultURL string `json:"result_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", err
	}

	return respData.ResultURL, nil
}
```

Now our test passes! It's just as satisfying as I promised.

In order to start using VCR, we need to initialize the Recorder and override `shortener.DefaultClient` at the beginning of the test:

```go
func TestShorten(t *testing.T) {
	r, err := recorder.New("fixtures/dev.to")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		require.NoError(t, r.Stop())
	}()

	if r.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	shortener.DefaultClient = r.GetDefaultClient()

	// ...
```

Run the test to generate `fixtures/dev.to.yaml` with details about the test's request and response. When we re-run the test, it uses the recorded response instead of reaching out to the server. Don't just take my word for it; turn off your computer's WiFi and re-run the tests!

You might also notice that the time it takes to run the test is relatively consistent since `go-vcr` records and replays the response duration. You can manually modify this field in the YAML to speed up the tests.


### Mocking Errors

To further demonstrate the benefits of this kind of testing, let's add another feature: retry after `429` response due to rate-limiting. Since we know the API's rate limit is per second, `Shorten` can automatically wait a second and retry if it receives a `429` response code.

I tried to reproduce this error using the API directly, but it seems like it responds with existing URLs from a cache before considering the rate limit. Rather than polluting the cache with bogus URLs, we can create our own mocks this time.

This is a simple process since we already have generated fixtures. After copy/pasting `fixtures/dev.to.yaml` to a new file, duplicate the successful request/response interaction and change first response's code from `200` to `429`. This fixture mimics a successful retry after rate-limiting failure.

The only difference between this test and the original test is the new fixture filename. The expected output is the same since `Shorten` should handle the error. This means we can throw the test in a loop to make it more dynamic:

```go
func TestShorten(t *testing.T) {
	fixtures := []string{
		"fixtures/dev.to",
		"fixtures/rate_limit",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			r, err := recorder.New(fixture)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				require.NoError(t, r.Stop())
			}()

			if r.Mode() != recorder.ModeRecordOnce {
				t.Fatal("Recorder should be in ModeRecordOnce")
			}

			shortener.DefaultClient = r.GetDefaultClient()

			shortened, err := shortener.Shorten("https://dev.to/calvinmclean")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if shortened != "https://cleanuri.com/7nPmQk" {
				t.Errorf("unexpected result: %v", shortened)
			}
		})
	}
}
```

Once again, the new test fails. This time due to the unhandled `429` response, so let's implement the new feature to pass the test. In order to maintain simplicity, our function handles the error using `time.Sleep` and a recursive call rather than dealing with the complexity of considering max retries and exponential backoffs:
```go
func Shorten(targetURL string) (string, error) {
	// ...
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests:
		time.Sleep(time.Second)
		return Shorten(targetURL)
	default:
		return "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	// ...
```

Now run the tests again and see them pass!

Take it a step further on your own and try adding a test for a bad request, which will occur when using an invalid URL like `my-fake-url`.

The full code for this example (and the bad request test) is available [on Github](https://github.com/calvinmclean/calvinmclean.github.io/blob/main/examples/go-vcr-testing/shortener_test.go).


## Conclusion

The benefits of VCR testing are clear from just this simple example, but they are even more impactful when dealing with complex applications where the requests and responses are unwieldy. Rather than dealing with tedious mocks or opting for no tests at all, I encourage you to give this a try in your own applications. If you already rely on integration tests, getting started with VCR is even easier since you already have real requests that can generate fixtures.

Check out more documentation and examples in the package's Github repository: https://github.com/dnaeon/go-vcr
