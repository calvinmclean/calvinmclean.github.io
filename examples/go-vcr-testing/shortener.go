package shortener

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

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

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests:
		time.Sleep(time.Second)
		return Shorten(targetURL)
	default:
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
