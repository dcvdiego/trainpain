package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/dcvdiego/trainpain-backend/internal/external"
	"github.com/dcvdiego/trainpain-backend/internal/repository/postgres"
	"go.uber.org/zap"
)

type RouteService struct {
	stationRepo *postgres.StationRepository
	routeRepo   *postgres.RouteRepository
	hspClient   *external.HSPClient
	logger      *zap.Logger
}

func NewRouteService(
	stationRepo *postgres.StationRepository,
	routeRepo *postgres.RouteRepository,
	hspClient *external.HSPClient,
	logger *zap.Logger,
) *RouteService {
	return &RouteService{
		stationRepo: stationRepo,
		routeRepo:   routeRepo,
		hspClient:   hspClient,
		logger:      logger,
	}
}

func (s *RouteService) GetReliability(ctx context.Context, query domain.RouteReliabilityQuery) (*domain.RouteReliabilityResponse, error) {
	// Get origin and destination stations
	origin, err := s.stationRepo.GetByCRS(ctx, query.OriginCRS)
	if err != nil {
		return nil, fmt.Errorf("origin station not found: %w", err)
	}

	destination, err := s.stationRepo.GetByCRS(ctx, query.DestinationCRS)
	if err != nil {
		return nil, fmt.Errorf("destination station not found: %w", err)
	}

	// Get or create route
	route, err := s.routeRepo.GetOrCreate(ctx, origin.ID, destination.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create route: %w", err)
	}

	// Set default values
	if query.AnalysisDays == 0 {
		query.AnalysisDays = 90
	}
	if query.DayFilter == "" {
		query.DayFilter = "all"
	}
	if query.TimeStart == "" {
		query.TimeStart = "00:00"
	}
	if query.TimeEnd == "" {
		query.TimeEnd = "23:59"
	}

	// Fetch data from HSP API
	metrics, err := s.fetchAndComputeMetrics(ctx, query, route.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute metrics: %w", err)
	}

	return &domain.RouteReliabilityResponse{
		Origin:      origin,
		Destination: destination,
		Metrics:     metrics,
		Query:       query,
	}, nil
}

func (s *RouteService) fetchAndComputeMetrics(
	ctx context.Context,
	query domain.RouteReliabilityQuery,
	routeID int,
) (*domain.RouteReliabilityMetrics, error) {
	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -query.AnalysisDays)

	// Convert times to HSP format (HHMM without colon)
	fromTime := convertTimeFormat(query.TimeStart)
	toTime := convertTimeFormat(query.TimeEnd)

	// HSP API requires days field with specific values: WEEKDAY, SATURDAY, or SUNDAY
	// For "all" days, we need to query all three and aggregate
	var allResponses []*external.HSPServiceMetricsResponse

	if query.DayFilter == "all" {
		// Query all three day types in parallel for better performance
		type result struct {
			resp    *external.HSPServiceMetricsResponse
			dayType string
			err     error
		}

		results := make(chan result, 3)
		dayTypes := []string{"WEEKDAY", "SATURDAY", "SUNDAY"}

		for _, dayType := range dayTypes {
			go func(dt string) {
				hspReq := external.HSPServiceMetricsRequest{
					FromLoc:  query.OriginCRS,
					ToLoc:    query.DestinationCRS,
					FromTime: fromTime,
					ToTime:   toTime,
					FromDate: startDate.Format("2006-01-02"),
					ToDate:   endDate.Format("2006-01-02"),
					Days:     &dt,
				}

				s.logger.Info("Querying HSP API",
					zap.String("from", query.OriginCRS),
					zap.String("to", query.DestinationCRS),
					zap.String("date_range", fmt.Sprintf("%s to %s", hspReq.FromDate, hspReq.ToDate)),
					zap.String("days", dt),
				)

				hspResp, err := s.hspClient.GetServiceMetrics(ctx, hspReq)
				results <- result{resp: hspResp, dayType: dt, err: err}
			}(dayType)
		}

		// Collect all results
		for i := 0; i < 3; i++ {
			res := <-results
			if res.err != nil {
				return nil, fmt.Errorf("HSP API error for %s: %w", res.dayType, res.err)
			}
			allResponses = append(allResponses, res.resp)
		}
	} else {
		// Single query for specific day filter
		days := convertDayFilter(query.DayFilter)
		hspReq := external.HSPServiceMetricsRequest{
			FromLoc:  query.OriginCRS,
			ToLoc:    query.DestinationCRS,
			FromTime: fromTime,
			ToTime:   toTime,
			FromDate: startDate.Format("2006-01-02"),
			ToDate:   endDate.Format("2006-01-02"),
			Days:     days,
		}

		s.logger.Info("Querying HSP API",
			zap.String("from", query.OriginCRS),
			zap.String("to", query.DestinationCRS),
			zap.String("date_range", fmt.Sprintf("%s to %s", hspReq.FromDate, hspReq.ToDate)),
			zap.String("days", *days),
		)

		hspResp, err := s.hspClient.GetServiceMetrics(ctx, hspReq)
		if err != nil {
			return nil, fmt.Errorf("HSP API error: %w", err)
		}
		allResponses = append(allResponses, hspResp)
	}

	// Aggregate and compute metrics from HSP responses
	metrics := computeMetricsFromHSP(allResponses, routeID, query, startDate, endDate)

	return metrics, nil
}

func convertDayFilter(filter string) *string {
	switch filter {
	case "weekday":
		s := "WEEKDAY"
		return &s
	case "saturday":
		s := "SATURDAY"
		return &s
	case "sunday":
		s := "SUNDAY"
		return &s
	default:
		// Default to WEEKDAY if not specified
		s := "WEEKDAY"
		return &s
	}
}

// convertTimeFormat converts time from "HH:MM" to "HHMM" format for HSP API
func convertTimeFormat(timeStr string) string {
	if timeStr == "" {
		return ""
	}
	// Remove colon from time (e.g., "07:00" -> "0700")
	return timeStr[0:2] + timeStr[3:5]
}

func computeMetricsFromHSP(
	hspResponses []*external.HSPServiceMetricsResponse,
	routeID int,
	query domain.RouteReliabilityQuery,
	startDate, endDate time.Time,
) *domain.RouteReliabilityMetrics {
	// Aggregate all services from multiple responses
	var allServices []external.HSPService
	for _, hspResp := range hspResponses {
		allServices = append(allServices, hspResp.Services...)
	}

	if len(allServices) == 0 {
		// No data available
		return &domain.RouteReliabilityMetrics{
			RouteID:               routeID,
			TimePeriod:            fmt.Sprintf("%s-%s", query.TimeStart, query.TimeEnd),
			StartTime:             &query.TimeStart,
			EndTime:               &query.TimeEnd,
			DayFilter:             query.DayFilter,
			AnalysisStartDate:     startDate,
			AnalysisEndDate:       endDate,
			TotalServicesAnalyzed: 0,
			ReliabilityScore:      0,
			ComputedAt:            time.Now(),
		}
	}

	// Extract and aggregate metrics from all HSP services
	// HSP provides metrics with different tolerance values (e.g., 0, 5, 10, 15 minutes)
	var (
		totalTolerance5    int
		totalNotTolerance5 int
		totalTolerance15   int
		totalTolerance30   int
	)

	// Aggregate metrics from all services
	for _, service := range allServices {
		for _, metric := range service.Metrics {
			if metric.ToleranceValue == 5 && metric.GlobalTolerance {
				totalTolerance5 += metric.NumTolerance
				totalNotTolerance5 += metric.NumNotTolerance
			}
			if metric.ToleranceValue == 15 && metric.GlobalTolerance {
				totalTolerance15 += metric.NumTolerance
			}
			if metric.ToleranceValue == 30 && metric.GlobalTolerance {
				totalTolerance30 += metric.NumTolerance
			}
		}
	}

	// Calculate percentages from aggregated totals
	totalServices := totalTolerance5 + totalNotTolerance5
	var onTimeRate, pct0To5Min, pct5To15Min, pct15To30Min float64

	if totalServices > 0 {
		onTimeRate = float64(totalTolerance5) / float64(totalServices) * 100
		pct0To5Min = onTimeRate
		pct5To15Min = (float64(totalTolerance15-totalTolerance5) / float64(totalServices)) * 100
		pct15To30Min = (float64(totalTolerance30-totalTolerance15) / float64(totalServices)) * 100
	}

	// Services delayed 30+ minutes
	pct30Plus := 100.0 - (pct0To5Min + pct5To15Min + pct15To30Min)

	// Estimate cancellation rate (simplified - in reality we'd need service details)
	// For MVP, assume 0 if we don't have this data
	cancellationRate := 0.0

	// Calculate average delay (rough estimation from distribution)
	avgDelay := (pct0To5Min * 2.5) + (pct5To15Min * 10) + (pct15To30Min * 22.5) + (pct30Plus * 45)
	avgDelay = avgDelay / 100.0

	// Calculate reliability score using the algorithm from the plan
	reliabilityScore := calculateReliabilityScore(onTimeRate, cancellationRate, avgDelay, pct30Plus)

	metrics := &domain.RouteReliabilityMetrics{
		RouteID:               routeID,
		TimePeriod:            fmt.Sprintf("%s-%s", query.TimeStart, query.TimeEnd),
		StartTime:             &query.TimeStart,
		EndTime:               &query.TimeEnd,
		DayFilter:             query.DayFilter,
		AnalysisStartDate:     startDate,
		AnalysisEndDate:       endDate,
		TotalServicesAnalyzed: totalServices,
		CancellationRate:      cancellationRate,
		OnTimeRate:            onTimeRate,
		AvgDelayMinutes:       avgDelay,
		MedianDelayMinutes:    int(avgDelay), // Simplified
		P95DelayMinutes:       int(pct30Plus * 0.5), // Rough estimate
		P99DelayMinutes:       int(pct30Plus * 0.7), // Rough estimate
		Pct0To5MinLate:        pct0To5Min,
		Pct5To15MinLate:       pct5To15Min,
		Pct15To30MinLate:      pct15To30Min,
		Pct30PlusMinLate:      pct30Plus,
		ReliabilityScore:      reliabilityScore,
		ComputedAt:            time.Now(),
	}

	return metrics
}

func calculateReliabilityScore(onTimeRate, cancellationRate, avgDelay, p95Delay float64) float64 {
	// Reliability Score Formula from the plan:
	// Score = (OnTime% × 0.4) + ((100 - Cancellation%) × 0.3) + ((100 - AvgDelay/30) × 100 × 0.2) + ((100 - P95Delay/60) × 100 × 0.1)

	score := (onTimeRate * 0.4) +
		((100 - cancellationRate) * 0.3) +
		((100 - (avgDelay / 30) * 100) * 0.2) +
		((100 - (p95Delay / 60) * 100) * 0.1)

	// Clamp between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
