package external

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

type HSPClient struct {
	baseURL  string
	username string
	password string
	client   *resty.Client
	limiter  *rate.Limiter
}

func NewHSPClient(baseURL, username, password string) *HSPClient {
	// Rate limiter: 5M requests per 4 weeks = ~178k per day
	// To be safe, limit to 100 requests per minute
	limiter := rate.NewLimiter(rate.Every(600*time.Millisecond), 10)

	return &HSPClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   resty.New().SetTimeout(30 * time.Second),
		limiter:  limiter,
	}
}

type HSPServiceMetricsRequest struct {
	FromLoc  string `json:"from_loc"`
	ToLoc    string `json:"to_loc"`
	FromTime string `json:"from_time"`
	ToTime   string `json:"to_time"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	Days     string `json:"days,omitempty"` // WEEKDAY, WEEKEND, or empty for all
}

type HSPServiceMetricsResponse struct {
	Services []HSPService `json:"Services"`
}

type HSPService struct {
	ServiceAttributesMetrics HSPServiceAttributesMetrics `json:"serviceAttributesMetrics"`
	Metrics                  []HSPMetric                  `json:"metrics"`
}

type HSPServiceAttributesMetrics struct {
	OriginLocation      string `json:"origin_location"`
	DestinationLocation string `json:"destination_location"`
	GbttPtd             string `json:"gbtt_ptd"`
	GbttPta             string `json:"gbtt_pta"`
	TocCode             string `json:"toc_code"`
	RIDs                []string `json:"rids,omitempty"`
}

type HSPMetric struct {
	ToleranceValue  int     `json:"tolerance_value"`
	NumNotTolerance int     `json:"num_not_tolerance"`
	NumTolerance    int     `json:"num_tolerance"`
	Percent         float64 `json:"percent"`
	GlobalTolerance bool    `json:"global_tolerance"`
}

type HSPServiceDetailsRequest struct {
	RID string `json:"rid"`
}

type HSPServiceDetailsResponse struct {
	ServiceAttributesDetails HSPServiceAttributesDetails `json:"serviceAttributesDetails"`
	Locations                []HSPLocation               `json:"locations"`
}

type HSPServiceAttributesDetails struct {
	DateOfService string `json:"date_of_service"`
	TocCode       string `json:"toc_code"`
	RID           string `json:"rid"`
	TrainUID      string `json:"trainUid"`
}

type HSPLocation struct {
	Location         string  `json:"location"`
	GbttPtd          *string `json:"gbtt_ptd,omitempty"`
	GbttPta          *string `json:"gbtt_pta,omitempty"`
	ActualTd         *string `json:"actual_td,omitempty"`
	ActualTa         *string `json:"actual_ta,omitempty"`
	LateCanc         bool    `json:"late_canc"`
	LateCancReason   *string `json:"late_canc_reason,omitempty"`
}

func (c *HSPClient) GetServiceMetrics(ctx context.Context, req HSPServiceMetricsRequest) (*HSPServiceMetricsResponse, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	var result HSPServiceMetricsResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(c.username, c.password).
		SetBody(req).
		SetResult(&result).
		Post(c.baseURL + "/api/v1/serviceMetrics")

	if err != nil {
		return nil, fmt.Errorf("HSP API request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HSP API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return &result, nil
}

func (c *HSPClient) GetServiceDetails(ctx context.Context, req HSPServiceDetailsRequest) (*HSPServiceDetailsResponse, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	var result HSPServiceDetailsResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(c.username, c.password).
		SetBody(req).
		SetResult(&result).
		Post(c.baseURL + "/api/v1/serviceDetails")

	if err != nil {
		return nil, fmt.Errorf("HSP API request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HSP API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return &result, nil
}
