package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	authLoginAPIPath = "api/rest-auth/login"
)

var ErrNoValidToken = fmt.Errorf("no valid token")
var ErrExpiredToken = fmt.Errorf("token is expired, please renew")
var ErrRateLimited = fmt.Errorf("rate limited response, please back off")

type authRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type authResponse struct {
	Key string `json:"key,omitempty"`
}

// GetAuthToken will obtain a new API authentication token
func (a *defaultApiClient) GetAuthToken(username, password string) (string, error) {
	c := a.getHttpClient()
	url := a.constructURL(authLoginAPIPath)
	data, err := json.Marshal(authRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.GetTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 429 {
		// back off though this is not documented.
		return "", ErrRateLimited
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("unable to obtain auth token: %v", string(respData))
	}
	var r authResponse
	err = json.Unmarshal(respData, &r)
	if err != nil {
		return "", err
	}

	return r.Key, nil
}

type inMemoryAuthTokenStorage struct {
	token  string
	expiry time.Time
	mu     sync.RWMutex
}

// NewInMemoryAuthTokenStorage will create a new in-memory auth token storage.
func NewInMemoryAuthTokenStorage() *inMemoryAuthTokenStorage {
	return &inMemoryAuthTokenStorage{}
}

// StoreToken will store this token in memory.
func (s *inMemoryAuthTokenStorage) StoreToken(token string, expiry time.Time) error {
	if token != "" {
		s.mu.Lock()
		s.token = token
		s.expiry = expiry
		s.mu.Unlock()
	}
	return nil
}

// GetCurrentToken will return the currently active token.
func (s *inMemoryAuthTokenStorage) GetCurrentToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.token == "" {
		return "", ErrNoValidToken
	}
	if s.expiry.Before(time.Now()) {
		return "", ErrExpiredToken
	}
	return s.token, nil
}

// RenewToken renews a token with given credentials
func (s *inMemoryAuthTokenStorage) RenewToken(client APIClient, username, password string) error {
	newToken, err := client.GetAuthToken(username, password)
	if err != nil {
		return err
	}
	err = s.StoreToken(newToken, time.Now().Add(time.Hour))
	if err != nil {
		return err
	}
	return nil
}

// SetRequestHeaders will get the request headers to set on authenticated routes
func SetRequestHeaders(req *http.Request, tokenStorage AuthTokenStorage) error {
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)
	if token, err := tokenStorage.GetCurrentToken(); err != nil {
		return err
	} else {
		req.Header.Set("Authorization", "Token "+token)
	}
	return nil
}
