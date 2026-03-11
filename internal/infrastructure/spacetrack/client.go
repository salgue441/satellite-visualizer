package spacetrack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/domain"
)

// Compile-time check that Client implements application.TLEProvider.
var _ application.TLEProvider = (*Client)(nil)

// Client fetches TLE data from the Space-Track.org API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	authCookie *http.Cookie
	mu         sync.Mutex // protects authCookie
}

// NewClient creates a new Space-Track client.
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: timeout,
			// Do not follow redirects automatically so we can capture cookies.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// FetchConstellation authenticates if needed, then fetches TLEs for a constellation.
func (c *Client) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error) {
	if c.username == "" || c.password == "" {
		return nil, domain.ErrAuthFailed
	}

	c.mu.Lock()
	needsAuth := c.authCookie == nil
	c.mu.Unlock()

	if needsAuth {
		if err := c.authenticate(ctx); err != nil {
			return nil, err
		}
	}

	tles, err := c.fetchTLEs(ctx, name)
	if err == errUnauthorized {
		// Session expired; clear cookie, re-authenticate, retry once.
		c.mu.Lock()
		c.authCookie = nil
		c.mu.Unlock()

		if authErr := c.authenticate(ctx); authErr != nil {
			return nil, authErr
		}
		return c.fetchTLEs(ctx, name)
	}
	return tles, err
}

// Available returns the list of curated constellation group names.
func (c *Client) Available() []string {
	return []string{
		"stations",     // ISS and other stations
		"starlink",     // SpaceX Starlink
		"gps-ops",      // GPS operational
		"oneweb",       // OneWeb
		"iridium-NEXT", // Iridium NEXT
		"galileo",      // Galileo
		"glo-ops",      // GLONASS operational
	}
}

// sentinel for 401 during fetch (not exported).
var errUnauthorized = fmt.Errorf("unauthorized")

// authenticate logs in to Space-Track and stores the session cookie.
func (c *Client) authenticate(ctx context.Context) error {
	form := url.Values{
		"identity": {c.username},
		"password": {c.password},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/ajaxauth/login",
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode != http.StatusOK {
		return domain.ErrAuthFailed
	}

	// Extract session cookie.
	for _, cookie := range resp.Cookies() {
		if cookie.Value != "" {
			c.mu.Lock()
			c.authCookie = cookie
			c.mu.Unlock()
			return nil
		}
	}

	return domain.ErrAuthFailed
}

// fetchTLEs performs the actual GET for TLE data.
func (c *Client) fetchTLEs(ctx context.Context, name string) ([]domain.TLE, error) {
	fetchURL := fmt.Sprintf(
		"%s/basicspacedata/query/class/tle_latest/ORDINAL/1/NORAD_CAT_ID/%%3E0/OBJECT_TYPE/PAYLOAD/orderby/NORAD_CAT_ID/format/3le/GROUP/%s",
		c.baseURL, name,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating fetch request: %w", err)
	}

	c.mu.Lock()
	if c.authCookie != nil {
		req.AddCookie(c.authCookie)
	}
	c.mu.Unlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching TLE data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errUnauthorized
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrConstellationNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return parseTLEResponse(string(body))
}

// parseTLEResponse parses Space-Track's 3-line TLE format.
func parseTLEResponse(body string) ([]domain.TLE, error) {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("%w: response too short", domain.ErrInvalidTle)
	}
	if len(lines)%3 != 0 {
		return nil, fmt.Errorf("%w: line count not divisible by 3", domain.ErrInvalidTle)
	}

	var tles []domain.TLE
	for i := 0; i < len(lines); i += 3 {
		name := strings.TrimSpace(lines[i])
		line1 := strings.TrimSpace(lines[i+1])
		line2 := strings.TrimSpace(lines[i+2])

		if len(line1) < 1 || line1[0] != '1' {
			return nil, fmt.Errorf("%w: expected line 1, got: %s", domain.ErrInvalidTle, line1)
		}
		if len(line2) < 1 || line2[0] != '2' {
			return nil, fmt.Errorf("%w: expected line 2, got: %s", domain.ErrInvalidTle, line2)
		}

		tles = append(tles, domain.TLE{
			Name:  name,
			Line1: line1,
			Line2: line2,
		})
	}

	return tles, nil
}
