package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	base string
	HTTP *http.Client
}

func NewClient(base string) *Client {
	return &Client{
		base: base,
		HTTP: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Get(path string, v any) error {
	resp, err := c.HTTP.Get(c.base + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s -> %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
