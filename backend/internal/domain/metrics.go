package domain

import "time"

type RouteReliabilityMetrics struct {
	ID                    int       `db:"id" json:"id"`
	RouteID               int       `db:"route_id" json:"route_id"`
	TimePeriod            string    `db:"time_period" json:"time_period"`
	StartTime             *string   `db:"start_time" json:"start_time,omitempty"`
	EndTime               *string   `db:"end_time" json:"end_time,omitempty"`
	DayFilter             string    `db:"day_filter" json:"day_filter"`
	AnalysisStartDate     time.Time `db:"analysis_start_date" json:"analysis_start_date"`
	AnalysisEndDate       time.Time `db:"analysis_end_date" json:"analysis_end_date"`
	TotalServicesAnalyzed int       `db:"total_services_analyzed" json:"total_services_analyzed"`
	CancellationRate      float64   `db:"cancellation_rate" json:"cancellation_rate"`
	OnTimeRate            float64   `db:"on_time_rate" json:"on_time_rate"`
	AvgDelayMinutes       float64   `db:"avg_delay_minutes" json:"avg_delay_minutes"`
	MedianDelayMinutes    int       `db:"median_delay_minutes" json:"median_delay_minutes"`
	P95DelayMinutes       int       `db:"p95_delay_minutes" json:"p95_delay_minutes"`
	P99DelayMinutes       int       `db:"p99_delay_minutes" json:"p99_delay_minutes"`
	Pct0To5MinLate        float64   `db:"pct_0_5_min_late" json:"pct_0_5_min_late"`
	Pct5To15MinLate       float64   `db:"pct_5_15_min_late" json:"pct_5_15_min_late"`
	Pct15To30MinLate      float64   `db:"pct_15_30_min_late" json:"pct_15_30_min_late"`
	Pct30PlusMinLate      float64   `db:"pct_30_plus_min_late" json:"pct_30_plus_min_late"`
	ReliabilityScore      float64   `db:"reliability_score" json:"reliability_score"`
	BestDayOfWeek         *string   `db:"best_day_of_week" json:"best_day_of_week,omitempty"`
	WorstDayOfWeek        *string   `db:"worst_day_of_week" json:"worst_day_of_week,omitempty"`
	MostCommonDelayReason *string   `db:"most_common_delay_reason" json:"most_common_delay_reason,omitempty"`
	ComputedAt            time.Time `db:"computed_at" json:"computed_at"`
}

// RouteReliabilityQuery represents a request for route reliability metrics
type RouteReliabilityQuery struct {
	OriginCRS      string `json:"origin_crs" binding:"required"`
	DestinationCRS string `json:"destination_crs" binding:"required"`
	TimeStart      string `json:"time_start,omitempty"`
	TimeEnd        string `json:"time_end,omitempty"`
	DayFilter      string `json:"day_filter,omitempty"` // weekday, weekend, all, monday, tuesday, etc.
	AnalysisDays   int    `json:"analysis_days,omitempty"` // 30, 60, 90, 180, 365
}

// RouteReliabilityResponse is the API response with station details
type RouteReliabilityResponse struct {
	Origin      *Station                 `json:"origin"`
	Destination *Station                 `json:"destination"`
	Metrics     *RouteReliabilityMetrics `json:"metrics"`
	Query       RouteReliabilityQuery    `json:"query"`
}
