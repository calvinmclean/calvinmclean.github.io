package shortener_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	shortener "go-vcr-testing-example"
)

func TestShorten(t *testing.T) {
	fixtures := []string{"fixtures/dev.to", "fixtures/rate_limit"}

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

func TestShortenTable(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectedErr string
	}{
		{
			"github",
			"http://github.com",
			"https://cleanuri.com/KGMrA7",
			"",
		},
		{
			"badrequest",
			"github.com",
			"",
			"unexpected response code: 400",
		},
		{
			"rate_limit",
			"https://dev.to/calvinmclean",
			"https://cleanuri.com/7nPmQk",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := recorder.New("fixtures/" + tt.name)
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

			shortened, err := shortener.Shorten(tt.url)
			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.expectedErr {
					t.Errorf("expected error, but got: %v", err)
				}
			}

			if shortened != tt.expected {
				t.Errorf("unexpected result: %v", shortened)
			}
		})
	}
}
