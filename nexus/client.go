package nexus

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
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

	// Load custom CA certificate if provided
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
		tlsConfig.InsecureSkipVerify = false
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     100,
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

// Upload uploads a file to the Nexus repository
func (c *Client) Upload(path string, data []byte) *UploadResult {
	result := &UploadResult{
		Path: path,
		Size: int64(len(data)),
	}

	start := time.Now()

	url := fmt.Sprintf("%s/repository/%s/%s", c.endpoint, c.repository, path)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		result.Error = err
		return result
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(data))

	resp, err := c.httpClient.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Bytes = int64(len(data))

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		result.Error = fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return result
}

// Download downloads a file from the Nexus repository
func (c *Client) Download(path string) *DownloadResult {
	result := &DownloadResult{
		Path: path,
	}

	start := time.Now()

	url := fmt.Sprintf("%s/repository/%s/%s", c.endpoint, c.repository, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = err
		return result
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}

	result.Size = int64(len(body))
	result.Bytes = int64(len(body))

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	return result
}

// Delete deletes a file from the Nexus repository
func (c *Client) Delete(path string) error {
	url := fmt.Sprintf("%s/repository/%s/%s", c.endpoint, c.repository, path)

	req, err := http.NewRequest("DELETE", url, nil)
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the client
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
