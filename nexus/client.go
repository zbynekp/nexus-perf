package nexus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/company/nexus-perf/config"
)

// Client represents a Nexus repository client
type Client struct {
	endpoint   string
	repository string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new Nexus client
func NewClient(cfg *config.Config) (*Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify,
	}

	if cfg.CAPath != "" {
		caCert, err := os.ReadFile(cfg.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// No overall Timeout on the client — large file transfers can take minutes.
	// Timeouts are applied only at connection/handshake/response-header level.
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			MaxIdleConns:        cfg.NumThreads,
			MaxIdleConnsPerHost: cfg.NumThreads,
			MaxConnsPerHost:     cfg.NumThreads,
			DialContext: (&net.Dialer{
				Timeout:   15 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}

	return &Client{
		endpoint:   strings.TrimRight(cfg.NexusEndpoint, "/"),
		repository: cfg.RepositoryName,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: httpClient,
	}, nil
}

// UploadResult contains upload result metrics
type UploadResult struct {
	Path       string
	Size       int64
	Duration   time.Duration
	Bytes      int64
	StatusCode int
	Error      error
}

// DownloadResult contains download result metrics
type DownloadResult struct {
	Path       string
	Size       int64
	Duration   time.Duration
	Bytes      int64
	StatusCode int
	Error      error
}

// Ping checks that the Nexus instance is reachable by hitting the status API.
func (c *Client) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/service/rest/v1/status", c.endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not build request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Nexus at %s: %w", c.endpoint, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Nexus status endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// CheckRepository verifies credentials via the authenticated status endpoint,
// then confirms the repository exists via a GET on its root path.
func (c *Client) CheckRepository(ctx context.Context) error {
	// Step 1: validate credentials using the auth-gated status endpoint.
	authURL := fmt.Sprintf("%s/service/rest/v1/status/check", c.endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	if err != nil {
		return fmt.Errorf("could not build request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid credentials (HTTP 401)")
	case http.StatusOK, http.StatusForbidden:
		// 200 = full access, 403 = limited-privilege user but creds are valid
	default:
		return fmt.Errorf("credential check returned HTTP %d", resp.StatusCode)
	}

	// Step 2: confirm the repository exists using the search API.
	// Nexus returns 422 when the repository name is unknown, and 200 (with
	// empty results) when it exists — even if the repository is empty.
	searchURL := fmt.Sprintf("%s/service/rest/v1/search?repository=%s&limit=1", c.endpoint, c.repository)
	req2, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return fmt.Errorf("could not build request: %w", err)
	}
	req2.SetBasicAuth(c.username, c.password)
	resp2, err := c.httpClient.Do(req2)
	if err != nil {
		return fmt.Errorf("repository check failed: %w", err)
	}
	_, _ = io.Copy(io.Discard, resp2.Body)
	resp2.Body.Close()
	switch resp2.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnprocessableEntity: // 422 — repository name not recognised
		return fmt.Errorf("repository %q not found (HTTP 422)", c.repository)
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid credentials (HTTP 401)")
	case http.StatusForbidden:
		return fmt.Errorf("access denied for repository %q (HTTP 403)", c.repository)
	default:
		return fmt.Errorf("repository check returned HTTP %d", resp2.StatusCode)
	}
}

// Upload uploads a file to the Nexus repository. The caller provides an
// io.Reader so data can be streamed without being fully buffered in memory.
func (c *Client) Upload(ctx context.Context, path string, size int64, body io.Reader) *UploadResult {
	result := &UploadResult{
		Path: path,
		Size: size,
	}

	start := time.Now()

	url := fmt.Sprintf("%s/repository/%s/%s", c.endpoint, c.repository, path)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		result.Error = err
		return result
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size

	resp, err := c.httpClient.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Bytes = size

	respBody, readErr := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		if readErr != nil {
			result.Error = fmt.Errorf("upload failed with status %d (response body unreadable: %w)", resp.StatusCode, readErr)
		} else {
			result.Error = fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
		}
	}

	return result
}

// Download downloads a file from the Nexus repository, copying the response
// body into dst (pass io.Discard to measure throughput without storing data).
func (c *Client) Download(ctx context.Context, path string, dst io.Writer) *DownloadResult {
	result := &DownloadResult{
		Path: path,
	}

	start := time.Now()

	url := fmt.Sprintf("%s/repository/%s/%s", c.endpoint, c.repository, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = err
		return result
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		result.Duration = time.Since(start)
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		result.Duration = time.Since(start)
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			result.Error = fmt.Errorf("download failed with status %d (response body unreadable: %w)", resp.StatusCode, readErr)
		} else {
			result.Error = fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
		}
		return result
	}

	n, err := io.Copy(dst, resp.Body)
	result.Duration = time.Since(start) // includes full body transfer
	result.Size = n
	result.Bytes = n
	if err != nil {
		result.Error = err
	}

	return result
}

// Delete deletes a file from the Nexus repository
func (c *Client) Delete(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s/repository/%s/%s", c.endpoint, c.repository, path)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("delete failed with status %d (response body unreadable: %w)", resp.StatusCode, readErr)
		}
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the client
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
