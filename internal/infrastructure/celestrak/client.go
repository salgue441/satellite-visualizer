package celestrak

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"satellite-visualizer/internal/application"
	"satellite-visualizer/internal/domain"
)

// Compile-time check that Client implements application.TLEProvider.
var _ application.TLEProvider = (*Client)(nil)

// Client fetches TLE data from the CelesTrak API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new CelesTrak client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// FetchConstellation retrieves TLEs for a named satellite group from CelesTrak.
func (c *Client) FetchConstellation(ctx context.Context, name string) ([]domain.TLE, error) {
	url := fmt.Sprintf("%s?GROUP=%s&FORMAT=tle", c.baseURL, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching TLE data: %w", err)
	}
	defer resp.Body.Close()

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

// parseTLEResponse parses CelesTrak's 3-line TLE format.
// Format: name line, TLE line 1, TLE line 2, repeated.
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
