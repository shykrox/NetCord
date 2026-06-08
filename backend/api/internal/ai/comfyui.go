package ai

import (
	"net/http"
	"strings"
	"time"
)

type ComfyUIClient struct {
	baseURL string
	http    *http.Client
}

func NewComfyUIClient(baseURL string) *ComfyUIClient {
	return &ComfyUIClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ComfyUIClient) Configured() bool {
	return c != nil && c.baseURL != ""
}
