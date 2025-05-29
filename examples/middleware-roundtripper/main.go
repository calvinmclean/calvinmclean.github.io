package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

var (
	serverLogger = slog.Default().With("source", "server")
	clientLogger = slog.Default().With("source", "client")
)

func main() {
	handler := RecoveryMiddleware(AuthMiddleware(LoggerMiddleware(CacheMiddleware(helloHandler))))

	// Bad: uses cache before auth
	// handler := CacheMiddleware(Authenticator(Logger(Recovery(helloHandler))))

	// Bad: logs the password
	// handler := Logger(Authenticator(Recovery(helloHandler)))

	http.HandleFunc("GET /hello", handler)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	time.Sleep(250 * time.Millisecond)

	client := &http.Client{
		// Bad: logs the password
		// Transport: AuthRoundTripper(LogRoundTripper(CacheRoundTripper(http.DefaultTransport))),

		// Bad: caches the password
		// Transport: LogRoundTripper(AuthRoundTripper(CacheRoundTripper(http.DefaultTransport))),

		// Good
		Transport: LogRoundTripper(CacheRoundTripper(AuthRoundTripper(http.DefaultTransport))),
	}

	// Disable Client cache and auth to show that server cache could skip auth
	noCacheClient := &http.Client{
		// Transport: LogRoundTripper(http.DefaultTransport),
		Transport: LogRoundTripper(AuthRoundTripper(http.DefaultTransport)),
	}

	doRequest(client)
	doRequest(client)
	doRequest(noCacheClient)
}

type responseWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func CacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	cache := map[string][]byte{}

	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := fmt.Sprintf("%s_%s", r.Method, r.URL.String())

		if cached, ok := cache[cacheKey]; ok {
			serverLogger.Info("SERVER using cached response")
			w.Header().Add("X-Server-Cached", "true")
			w.Write(cached)
			w.WriteHeader(http.StatusOK)
			return
		}

		var buf bytes.Buffer
		writer := &responseWriter{ResponseWriter: w, buffer: &buf}

		next(writer, r)

		cache[cacheKey] = buf.Bytes()
	}
}

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		serverLogger.Info(fmt.Sprintf("SERVER Request: %s %s %v", r.Method, r.URL.Path, r.Header))
		var buf bytes.Buffer
		writer := &responseWriter{ResponseWriter: w, buffer: &buf}
		next(writer, r)
		serverLogger.Info(fmt.Sprintf("SERVER Response: %s %s %d %s", r.Method, r.URL.Path, writer.statusCode, time.Since(start)))
	}
}

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
		next(w, r)
	}
}

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				serverLogger.Error(fmt.Sprintf("SERVER panic: %v", err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

func SequentialMiddleware(middleware ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for _, mw := range slices.Backward(middleware) {
			final = mw(final)
		}
		return final
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// panic("Something went wrong!")
	fmt.Fprintln(w, "Hello, World!")
	w.WriteHeader(http.StatusOK)
}

func doRequest(client *http.Client) {
	resp, err := client.Get("http://localhost:8080/hello")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Status)
	fmt.Println(string(body))
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (rt RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	if rt == nil {
		return http.DefaultTransport.RoundTrip(r)
	}
	return rt(r)
}

func LogRoundTripper(next http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		clientLogger.Info(fmt.Sprintf("CLIENT Request: %s %s %v", r.Method, r.URL, r.Header))
		start := time.Now()
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

func AuthRoundTripper(next http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		r.Header.Add("Authorization", "password")
		return next.RoundTrip(r)
	})
}

type cachedResponse struct {
	resp http.Response
	body []byte
}

func newCachedResponse(resp *http.Response) cachedResponse {
	cachedResp := cachedResponse{
		resp: *resp,
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	cachedResp.body = bodyBytes

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return cachedResp
}

func (cr cachedResponse) Response() *http.Response {
	cr.resp.Body = io.NopCloser(bytes.NewReader(cr.body))
	cr.resp.ContentLength = int64(len(cr.body))
	cr.resp.Header.Add("X-Client-Cached", "true")
	return &cr.resp
}

func CacheRoundTripper(next http.RoundTripper) http.RoundTripper {
	cache := map[string]cachedResponse{}

	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		cacheKey := fmt.Sprintf("%s_%s", r.Method, r.URL.String())

		// Detect password for demo
		password := r.Header.Get("Authorization")
		if password != "" {
			clientLogger.Warn("are you sure that you want to cache the password?", "password", password)
		}

		cachedResponse, ok := cache[cacheKey]
		if ok {
			clientLogger.Info("CLIENT using cached response")
			return cachedResponse.Response(), nil
		}

		resp, err := next.RoundTrip(r)
		cache[cacheKey] = newCachedResponse(resp)
		return resp, err
	})
}

func NewClientWithRoundTrippers(rts ...func(http.RoundTripper) http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: SequentialRoundTripper(rts...),
	}
}

func SequentialRoundTripper(rts ...func(http.RoundTripper) http.RoundTripper) http.RoundTripper {
	result := http.DefaultTransport

	for _, rt := range slices.Backward(rts) {
		result = RoundTripperFunc(rt(result).RoundTrip)
	}

	return result
}
