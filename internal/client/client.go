package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultAuthPath   = "/oauth2/v1/token"
	tokenExpiryJitter = 30 * time.Second
)

// Config represents the inputs required to build the API client.
type Config struct {
	Host       string
	Username   string
	Password   string
	ClientCert string
	ClientKey  string
	CACert     string
}

// Client wraps an HTTP client with mTLS, Basic Auth login and Bearer token refresh.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
	authPath    string
}

// tokenResponse captures common fields returned by token endpoints.
type tokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresAt   string `json:"expires_at"`
}

// New constructs a Client configured for mTLS and token-based authentication.
func New(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, errors.New("host must be provided")
	}
	if cfg.Username == "" {
		return nil, errors.New("username must be provided")
	}
	if cfg.Password == "" {
		return nil, errors.New("password must be provided")
	}
	if cfg.ClientCert == "" {
		return nil, errors.New("client certificate must be provided")
	}
	if cfg.ClientKey == "" {
		return nil, errors.New("client private key must be provided")
	}

	cert, err := tls.X509KeyPair([]byte(cfg.ClientCert), []byte(cfg.ClientKey))
	if err != nil {
		return nil, fmt.Errorf("loading client certificate/key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if strings.TrimSpace(cfg.CACert) != "" {
		rootPool := x509.NewCertPool()
		if ok := rootPool.AppendCertsFromPEM([]byte(cfg.CACert)); !ok {
			return nil, errors.New("failed to append provided CA certificate")
		}
		tlsConfig.RootCAs = rootPool
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	baseURL := strings.TrimRight(cfg.Host, "/")

	return &Client{
		baseURL:    baseURL,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: &http.Client{Transport: transport, Timeout: 30 * time.Second},
		authPath:   defaultAuthPath,
	}, nil
}

// do performs an authenticated HTTP request, refreshing the bearer token as needed.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	if err := c.ensureToken(ctx, false); err != nil {
		return err
	}

	attempt := 0
retry:
	attempt++

	var reqBody io.Reader
	if body != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reqBody = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If unauthorized, try refreshing the token once.
	if resp.StatusCode == http.StatusUnauthorized && attempt == 1 {
		if err := c.ensureToken(ctx, true); err != nil {
			return err
		}
		goto retry
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(b)))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// ensureToken guarantees that a valid bearer token exists. Force refresh when reauth is true.
func (c *Client) ensureToken(ctx context.Context, reauth bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if !reauth && c.token != "" && now.Add(tokenExpiryJitter).Before(c.tokenExpiry) {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.authPath, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var tokenRes tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return fmt.Errorf("decoding token response: %w", err)
	}

	token := tokenRes.Token
	if token == "" {
		token = tokenRes.AccessToken
	}
	if token == "" {
		return errors.New("token response did not include a token")
	}

	expiry := now.Add(5 * time.Minute)
	switch {
	case tokenRes.ExpiresIn > 0:
		expiry = now.Add(time.Duration(tokenRes.ExpiresIn) * time.Second)
	case tokenRes.ExpiresAt != "":
		if t, err := time.Parse(time.RFC3339, tokenRes.ExpiresAt); err == nil {
			expiry = t
		}
	}

	c.token = token
	c.tokenExpiry = expiry

	return nil
}
