package client

import (
	"net/http"
	"os"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		http:    &http.Client{},
	}
}

func BaseURL() string {
	if v := os.Getenv("PATERNA_API"); v != "" {
		return v
	}
	return "http://localhost:8080"
}
