package httpclient

import (
    "io"
    "net/http"
    "time"
)

type HttpClient struct {
    client *http.Client
}

func NewHttpClient() *HttpClient {
    return &HttpClient{
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (h *HttpClient) Get(url string) (*http.Response, error) {
    resp, err := h.client.Get(url)
    if err != nil {
        return nil, err
    }
    return resp, nil
}

func (h *HttpClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
    resp, err := h.client.Post(url, contentType, body)
    if err != nil {
        return nil, err
    }
    return resp, nil
}