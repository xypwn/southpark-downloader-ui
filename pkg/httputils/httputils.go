package httputils

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"southpark-downloader-ui/pkg/ioutils"
)

func GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func GetBodyWithContext(ctx context.Context, url string) ([]byte, error) {
	resp, err := GetWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get '%v': HTTP error: %v", url, resp.Status)
	}

	body, err := io.ReadAll(ioutils.NewCtxReader(ctx, resp.Body))
	if err != nil {
		return nil, fmt.Errorf("get '%v': io.ReadAll: %w", url, err)
	}

	return body, nil
}
