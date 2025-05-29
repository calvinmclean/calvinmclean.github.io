package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestBad_ClientAndServerLogPassword(t *testing.T) {
	serverHandler := NewTestLogHandler()
	clientHandler := NewTestLogHandler()
	serverLogger = slog.New(serverHandler)
	clientLogger = slog.New(clientHandler)

	// The server runs the LoggerMiddleware before the AuthMiddleware, which performs auth and then sanitizes the request
	handler := SequentialMiddleware(
		LoggerMiddleware,
		RecoveryMiddleware,
		AuthMiddleware,
		CacheMiddleware,
	)(helloHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	// The client adds authentication to the request before logging and caching it
	client := NewClientWithRoundTrippers(
		AuthRoundTripper,
		LogRoundTripper,
		CacheRoundTripper,
	)

	body, resp, err := doTestRequest(client, server.URL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectations{
		StatusCode: http.StatusOK,
		Body:       "Hello, World!\n",
	}.assert(t, body, resp)

	if !serverHandler.loggedPassword {
		t.Error("expected server password log")
	}
	if !clientHandler.loggedPassword {
		t.Error("expected client password log")
	}
}

func TestBad_ServerCacheExposesAccess(t *testing.T) {
	// The server will cache responses before checking authentication. After caching a response, the next request
	// will receive this response even if it does not authenticate successfully
	handler := SequentialMiddleware(
		CacheMiddleware,
		AuthMiddleware,
		LoggerMiddleware,
		RecoveryMiddleware,
	)(helloHandler)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("CreateInitialRequest_WithAuth", func(t *testing.T) {
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

type TestLogHandler struct {
	slog.Handler
	records        []slog.Record
	loggedPassword bool
}

func NewTestLogHandler() *TestLogHandler {
	return &TestLogHandler{Handler: slog.Default().Handler()}
}

func (h *TestLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if strings.Contains(record.Message, "password") {
		h.loggedPassword = true
	}
	h.records = append(h.records, record)
	return h.Handler.Handle(ctx, record)
}

func doTestRequest(client *http.Client, url string) (string, *http.Response, error) {
	r, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
	resp, err := client.Do(r)
	if err != nil {
		return "", resp, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp, err
	}

	return string(body), resp, nil
}

type expectations struct {
	StatusCode      int
	Body            string
	ClientCacheUsed bool
	ServerCacheUsed bool
}

func (e expectations) assert(t *testing.T, body string, resp *http.Response) {
	if resp.StatusCode != e.StatusCode {
		t.Errorf("unexpected status: %d", resp.StatusCode)
	}
	if body != e.Body {
		t.Errorf("unexpected body: %q", body)
	}

	expectedClientCache := ""
	if e.ClientCacheUsed {
		expectedClientCache = "true"
	}

	if resp.Header.Get("X-Client-Cached") != expectedClientCache {
		t.Errorf("X-Client-Cached did not match %q", expectedClientCache)
	}

	expectedServerCache := ""
	if e.ServerCacheUsed {
		expectedServerCache = "true"
	}
	if resp.Header.Get("X-Server-Cached") != expectedServerCache {
		t.Errorf("X-Server-Cached did not match %q", expectedServerCache)
	}
}
