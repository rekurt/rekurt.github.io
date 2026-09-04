package githubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/buildinfo"
)

const (
	defaultBaseURL = "https://api.github.com"
	maxAttempts    = 3
	maxBodyBytes   = 8 << 20
	maxErrorBytes  = 8 << 10
)

type cacheEntry struct {
	etag string
	body []byte
	link string
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("GitHub API returned %d: %s", e.StatusCode, e.Body)
}

func New(baseURL, token string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: httpClient,
		cache:      make(map[string]cacheEntry),
	}
}

func (c *Client) get(ctx context.Context, path string, output any) (http.Header, error) {
	requestURL := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		requestURL = c.baseURL + path
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("create GitHub request: %w", err)
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		request.Header.Set("User-Agent", "rekurt-portfolio-sync/"+buildinfo.Version)
		if c.token != "" {
			request.Header.Set("Authorization", "Bearer "+c.token)
		}
		c.mu.Lock()
		cached, cachedOK := c.cache[requestURL]
		c.mu.Unlock()
		if cachedOK && cached.etag != "" {
			request.Header.Set("If-None-Match", cached.etag)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("request GitHub %s: %w", requestURL, err)
		}

		if response.StatusCode == http.StatusNotModified && cachedOK {
			response.Body.Close()
			headers := response.Header.Clone()
			if headers.Get("Link") == "" {
				headers.Set("Link", cached.link)
			}
			if err := json.Unmarshal(cached.body, output); err != nil {
				return nil, fmt.Errorf("decode cached GitHub response %s: %w", requestURL, err)
			}
			return headers, nil
		}

		if isRetryable(response.StatusCode) && attempt+1 < maxAttempts {
			delay := retryDelay(response, attempt)
			io.Copy(io.Discard, io.LimitReader(response.Body, maxErrorBytes))
			response.Body.Close()
			if err := wait(ctx, delay); err != nil {
				return nil, err
			}
			continue
		}

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			body, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBytes))
			response.Body.Close()
			return nil, &HTTPError{StatusCode: response.StatusCode, Body: strings.TrimSpace(string(body))}
		}

		body, err := io.ReadAll(io.LimitReader(response.Body, maxBodyBytes+1))
		response.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read GitHub response %s: %w", requestURL, err)
		}
		if len(body) > maxBodyBytes {
			return nil, fmt.Errorf("GitHub response %s exceeds %d bytes", requestURL, maxBodyBytes)
		}
		if err := json.Unmarshal(body, output); err != nil {
			return nil, fmt.Errorf("decode GitHub response %s: %w", requestURL, err)
		}
		if etag := response.Header.Get("ETag"); etag != "" {
			c.mu.Lock()
			c.cache[requestURL] = cacheEntry{etag: etag, body: body, link: response.Header.Get("Link")}
			c.mu.Unlock()
		}
		return response.Header.Clone(), nil
	}

	return nil, errors.New("GitHub request exhausted retries")
}

func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func retryDelay(response *http.Response, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(response.Header.Get("Retry-After")); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 100 * time.Millisecond
}

func wait(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isNotFound(err error) bool {
	var apiError *HTTPError
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound
}
