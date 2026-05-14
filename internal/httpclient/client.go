package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

var Client *http.Client

func Init(maxIdle int) {
	Client = &http.Client{Timeout: 30 * time.Second, Transport: &http.Transport{MaxIdleConns: maxIdle, MaxIdleConnsPerHost: maxIdle}}
}
func Get(url string, headers map[string]string) ([]byte, int, error) {
	return do(http.MethodGet, url, headers, nil)
}
func Post(url string, headers map[string]string, body []byte) ([]byte, int, error) {
	return do(http.MethodPost, url, headers, body)
}
func do(method, url string, headers map[string]string, body []byte) ([]byte, int, error) {
	if Client == nil {
		Init(100)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}
