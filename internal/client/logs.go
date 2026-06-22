package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) StreamLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	url := c.baseURL + "/containers/" + id + "/logs/stream"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stream logs request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("stream logs (status %d): %s", resp.StatusCode, body)
	}

	return resp.Body, nil
}
