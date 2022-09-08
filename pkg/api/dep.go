package api

import (
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	contentType = "application/json"
)

type APIClient interface {
	// GetBaseURL get the base fireboard api url
	GetBaseURL() string
	// SetBaseURL update the base fireboard api url
	SetBaseURL(url string)
	// SetTimeout sets the timeout configuration value
	SetTimeout(timeout time.Duration)
	// GetTimeout get the timeout configuration value
	GetTimeout() time.Duration

	// GetAuthToken will obtain a new authentication token
	GetAuthToken(username, password string) (string, error)

	// ListDevices will list all devices
	ListDevices() (ListDevicesResponse, error)
	// GetDevice will get a single device information
	GetDevice(deviceUUID string) (*DevicePropertiesResponse, error)
	// GetRealTimeDeviceTemperature will get the current real-time device temperatures. Only valid for active sessions.
	GetRealTimeDeviceTemperature(deviceUUID string) (*DevicePropertiesResponse, error)
	// GetRealTimeDeviceDriveData will get the current real-time device drive data. Only valid for active sessions.
	GetRealTimeDeviceDriveData(deviceUUID string) (*DevicePropertiesResponse, error)

	// ListAllSessions list all sessions
	ListAllSessions() (SessionsListResponse, error)
	// GetSession will get a specific session
	GetSession(sessionID int64) (*SessionGetResponse, error)
	// GetSessionChartData will get the session chart data
	GetSessionChartData(sessionID int64) (SessionChartResponse, error)
}

// AuthTokenStorage implements a storage mechanism for the auth token.
type AuthTokenStorage interface {
	StoreToken(token string, expiry time.Time) error
	GetCurrentToken() (string, error)
	RenewToken(client APIClient, username, password string) error
}

type defaultApiClient struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
	authStore  AuthTokenStorage

	mu sync.RWMutex
}

// NewDefaultAPIClient returns a new defaultApiClient
func NewDefaultAPIClient() *defaultApiClient {
	baseURL := "https://fireboard.io"
	if val, ok := os.LookupEnv("FIREBOARD_API_URL"); ok && val != "" {
		baseURL = val
	}
	timeout := time.Second * 10
	if val, ok := os.LookupEnv("FIREBOARD_API_TIMEOUT_MILLIS"); ok && val != "" {
		if d, err := time.ParseDuration(val); err == nil && d.Milliseconds() > 0 {
			timeout = d
		}
	}

	return &defaultApiClient{
		baseURL: baseURL,
		timeout: timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		authStore: NewInMemoryAuthTokenStorage(),
	}
}

func (a *defaultApiClient) getHttpClient() *http.Client {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.httpClient
}

func (a *defaultApiClient) constructURL(path string) string {
	u, err := url.Parse(a.GetBaseURL())
	if err != nil {
		panic(err)
	}
	u.Path = path
	return u.String()
}

func (a *defaultApiClient) GetBaseURL() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.baseURL
}

// SetBaseURL update the base fireboard api url
func (a *defaultApiClient) SetBaseURL(url string) {
	a.mu.Lock()
	a.baseURL = url
	a.mu.Unlock()
}

// GetTimeout get the timeout configuration value
func (a *defaultApiClient) GetTimeout() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.timeout
}

// SetTimeout sets the timeout configuration value
func (a *defaultApiClient) SetTimeout(timeout time.Duration) {
	a.mu.Lock()
	a.timeout = timeout
	a.httpClient = &http.Client{
		Timeout: timeout,
	}
	a.mu.Unlock()
}
