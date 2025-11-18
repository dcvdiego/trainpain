package external

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

type TfLClient struct {
	baseURL string
	appID   string
	appKey  string
	client  *resty.Client
	limiter *rate.Limiter
}

func NewTfLClient(appID, appKey string) *TfLClient {
	// Rate limiter: Be conservative, 10 requests per minute
	limiter := rate.NewLimiter(rate.Every(6*time.Second), 5)

	return &TfLClient{
		baseURL: "https://api.tfl.gov.uk",
		appID:   appID,
		appKey:  appKey,
		client:  resty.New().SetTimeout(30 * time.Second),
		limiter: limiter,
	}
}

type TfLLineStatus struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ModeName   string `json:"modeName"`
	Disruptions []TfLDisruption `json:"disruptions,omitempty"`
	LineStatuses []TfLStatus `json:"lineStatuses"`
}

type TfLDisruption struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Created     string `json:"created"`
}

type TfLStatus struct {
	StatusSeverity int    `json:"statusSeverity"`
	StatusSeverityDescription string `json:"statusSeverityDescription"`
	Reason string `json:"reason,omitempty"`
}

func (c *TfLClient) GetLineStatus(ctx context.Context, lineIDs []string) ([]TfLLineStatus, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	var result []TfLLineStatus

	// Build line IDs string
	lineIDsStr := ""
	for i, id := range lineIDs {
		if i > 0 {
			lineIDsStr += ","
		}
		lineIDsStr += id
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"app_id":  c.appID,
			"app_key": c.appKey,
		}).
		SetResult(&result).
		Get(fmt.Sprintf("%s/Line/%s/Status", c.baseURL, lineIDsStr))

	if err != nil {
		return nil, fmt.Errorf("TfL API request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("TfL API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return result, nil
}

// GetDisruptionsByMode gets disruptions for specific modes (tube, overground, elizabeth-line, dlr)
func (c *TfLClient) GetDisruptionsByMode(ctx context.Context, modes []string) ([]TfLLineStatus, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	var result []TfLLineStatus

	// Build modes string
	modesStr := ""
	for i, mode := range modes {
		if i > 0 {
			modesStr += ","
		}
		modesStr += mode
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"app_id":  c.appID,
			"app_key": c.appKey,
		}).
		SetResult(&result).
		Get(fmt.Sprintf("%s/Line/Mode/%s/Disruption", c.baseURL, modesStr))

	if err != nil {
		return nil, fmt.Errorf("TfL API request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("TfL API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return result, nil
}
