package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	defaultAuthPath   = "/oauth2/v1/token"
	tokenExpiryJitter = 30 * time.Second
)

// Config represents the inputs required to build the API client.
type Config struct {
	Host                string
	ClientID            string
	ClientSecret        string
	ClientCert          string
	ClientKey           string
	ClientKeyPassphrase string
	CACert              string
	AuthPath            string
}

// Client wraps an HTTP client with mTLS, Basic Auth login and Bearer token refresh.
type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
	authPath    string
}

// tokenResponse captures the minimal fields returned by the token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// HTTPClient exposes the minimal surface needed by resources to perform
// authenticated JSON requests. It lets resource packages remain decoupled
// from the underlying authentication and mTLS setup while still sharing the
// refreshable bearer token logic.
type HTTPClient interface {
	Request(ctx context.Context, method, path string, body any, out any) error
	Get(ctx context.Context, path string, out any) error
	Post(ctx context.Context, path string, body any, out any) error
	Put(ctx context.Context, path string, body any, out any) error
	Delete(ctx context.Context, path string) error
}

var _ HTTPClient = (*Client)(nil)

// New constructs a Client configured for mTLS and token-based authentication.
func New(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, errors.New("host must be provided")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("client id must be provided")
	}
	if cfg.ClientSecret == "" {
		return nil, errors.New("client secret must be provided")
	}
	if cfg.ClientCert == "" {
		return nil, errors.New("client certificate must be provided")
	}
	if cfg.ClientKey == "" {
		return nil, errors.New("client private key must be provided")
	}

	keyBytes := []byte(cfg.ClientKey)
	if strings.TrimSpace(cfg.ClientKeyPassphrase) != "" {
		parsedKey, err := ssh.ParseRawPrivateKeyWithPassphrase(keyBytes, []byte(cfg.ClientKeyPassphrase))
		if err != nil {
			return nil, fmt.Errorf("decrypting private key with passphrase: %w", err)
		}
		der, err := x509.MarshalPKCS8PrivateKey(parsedKey)
		if err != nil {
			return nil, fmt.Errorf("marshalling decrypted private key: %w", err)
		}
		keyBytes = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	}

	cert, err := tls.X509KeyPair([]byte(cfg.ClientCert), keyBytes)
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
	authPath := defaultAuthPath
	if strings.TrimSpace(cfg.AuthPath) != "" {
		authPath = cfg.AuthPath
	}

	return &Client{
		baseURL:      baseURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   &http.Client{Transport: transport, Timeout: 30 * time.Second},
		authPath:     authPath,
	}, nil
}

// Request executes an authenticated request and decodes JSON into out.
func (c *Client) Request(ctx context.Context, method, path string, body any, out any) error {
	return c.do(ctx, method, path, body, out)
}

// Get performs an authenticated GET request.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post performs an authenticated POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// Put performs an authenticated PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

// Delete performs an authenticated DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
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
	req.SetBasicAuth(c.clientID, c.clientSecret)
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

	if tokenRes.AccessToken == "" {
		return errors.New("token response did not include access_token")
	}

	expiry := now.Add(30 * time.Minute)
	if tokenRes.ExpiresIn > 0 {
		expiry = now.Add(time.Duration(tokenRes.ExpiresIn) * time.Second)
	}

	c.token = tokenRes.AccessToken
	c.tokenExpiry = expiry

	return nil
}
