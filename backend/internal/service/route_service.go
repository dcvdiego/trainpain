package service

import (
	"context"
	"fmt"
	"strconv"
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
	// Default to morning peak hours to avoid overwhelming the HSP API
	// Querying all 24 hours causes 502 errors on busy routes
	if query.TimeStart == "" {
		query.TimeStart = "07:00"
	}
	if query.TimeEnd == "" {
		query.TimeEnd = "09:00"
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

	// Split large date ranges into 1-day chunks to avoid HSP API timeouts
	// The HSP API backend struggles even with 7-day queries on very busy routes
	const maxDaysPerQuery = 1
	var allResponses []*external.HSPServiceMetricsResponse

	// Calculate date chunks
	dateChunks := splitDateRange(startDate, endDate, maxDaysPerQuery)

	// Filter chunks by specific day of week if needed
	dateChunks = filterChunksByDayOfWeek(dateChunks, query.DayFilter)

	s.logger.Info("Querying HSP API",
		zap.String("from", query.OriginCRS),
		zap.String("to", query.DestinationCRS),
		zap.Int("total_days", query.AnalysisDays),
		zap.Int("chunks", len(dateChunks)),
		zap.String("day_filter", query.DayFilter),
	)

	// HSP API requires days field with specific values: WEEKDAY, SATURDAY, or SUNDAY
	// For "all" days, we need to query all three and aggregate
	if query.DayFilter == "all" {
		// Query all three day types in parallel for each date chunk
		dayTypes := []string{"WEEKDAY", "SATURDAY", "SUNDAY"}

		for _, chunk := range dateChunks {
			type result struct {
				resp    *external.HSPServiceMetricsResponse
				dayType string
				err     error
			}

			results := make(chan result, 3)

			for _, dayType := range dayTypes {
				go func(dt string, chunkStart, chunkEnd time.Time) {
					hspReq := external.HSPServiceMetricsRequest{
						FromLoc:  query.OriginCRS,
						ToLoc:    query.DestinationCRS,
						FromTime: fromTime,
						ToTime:   toTime,
						FromDate: chunkStart.Format("2006-01-02"),
						ToDate:   chunkEnd.Format("2006-01-02"),
						Days:     &dt,
						// Note: tolerance parameter causes 30+ second timeouts, removed
					}

					s.logger.Debug("Querying HSP API chunk",
						zap.String("days", dt),
						zap.String("date_range", fmt.Sprintf("%s to %s", hspReq.FromDate, hspReq.ToDate)),
						zap.String("time_window", fmt.Sprintf("%s to %s", hspReq.FromTime, hspReq.ToTime)),
					)

					hspResp, err := s.hspClient.GetServiceMetrics(ctx, hspReq)
					if err == nil {
						s.logger.Debug("HSP API response",
							zap.String("days", dt),
							zap.Int("services_returned", len(hspResp.Services)),
						)
					}
					results <- result{resp: hspResp, dayType: dt, err: err}
				}(dayType, chunk.start, chunk.end)
			}

			// Collect results for this chunk
			for i := 0; i < 3; i++ {
				res := <-results
				if res.err != nil {
					return nil, fmt.Errorf("HSP API error for %s: %w", res.dayType, res.err)
				}
				allResponses = append(allResponses, res.resp)
			}
		}
	} else {
		// Single day type - process chunks in parallel for better performance
		days := convertDayFilter(query.DayFilter)

		// Use worker pool pattern to parallelize API calls
		// Limit concurrency to avoid overwhelming the API (3 workers = sweet spot)
		// HSP API returns 503 if hit too hard, so keep concurrency low
		const maxConcurrency = 3
		semaphore := make(chan struct{}, maxConcurrency)

		type result struct {
			resp  *external.HSPServiceMetricsResponse
			index int
			err   error
		}

		results := make(chan result, len(dateChunks))

		// Launch goroutines for each chunk
		for i, chunk := range dateChunks {
			go func(idx int, chunkStart, chunkEnd time.Time) {
				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }() // Release semaphore

				hspReq := external.HSPServiceMetricsRequest{
					FromLoc:  query.OriginCRS,
					ToLoc:    query.DestinationCRS,
					FromTime: fromTime,
					ToTime:   toTime,
					FromDate: chunkStart.Format("2006-01-02"),
					ToDate:   chunkEnd.Format("2006-01-02"),
					Days:     days,
					// Note: tolerance parameter causes 30+ second timeouts, removed
				}

				s.logger.Debug("Querying HSP API chunk",
					zap.String("days", *days),
					zap.String("date_range", fmt.Sprintf("%s to %s", hspReq.FromDate, hspReq.ToDate)),
					zap.String("time_window", fmt.Sprintf("%s to %s", hspReq.FromTime, hspReq.ToTime)),
				)

				hspResp, err := s.hspClient.GetServiceMetrics(ctx, hspReq)
				if err == nil {
					s.logger.Debug("HSP API response",
						zap.String("days", *days),
						zap.Int("services_returned", len(hspResp.Services)),
					)
				}

				results <- result{resp: hspResp, index: idx, err: err}
			}(i, chunk.start, chunk.end)
		}

		// Collect results in order
		responseMap := make(map[int]*external.HSPServiceMetricsResponse)
		for i := 0; i < len(dateChunks); i++ {
			res := <-results
			if res.err != nil {
				return nil, fmt.Errorf("HSP API error: %w", res.err)
			}
			responseMap[res.index] = res.resp
		}

		// Add responses in original order
		for i := 0; i < len(dateChunks); i++ {
			allResponses = append(allResponses, responseMap[i])
		}
	}

	// Aggregate and compute metrics from HSP responses
	metrics, err := computeMetricsFromHSP(ctx, s.hspClient, allResponses, routeID, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to compute metrics: %w", err)
	}

	return metrics, nil
}

type dateChunk struct {
	start time.Time
	end   time.Time
}

func splitDateRange(start, end time.Time, maxDays int) []dateChunk {
	var chunks []dateChunk
	current := start

	for current.Before(end) || current.Equal(end) {
		// For 1-day chunks, chunkEnd should be same as start
		// For N-day chunks, chunkEnd should be N-1 days after start
		chunkEnd := current.AddDate(0, 0, maxDays-1)
		if chunkEnd.After(end) {
			chunkEnd = end
		}
		chunks = append(chunks, dateChunk{start: current, end: chunkEnd})
		current = chunkEnd.AddDate(0, 0, 1) // Move to next day after chunk
	}

	return chunks
}

func filterChunksByDayOfWeek(chunks []dateChunk, dayFilter string) []dateChunk {
	// If "all" or "weekday", return all chunks
	// "weekday" maps to WEEKDAY HSP API type which includes Mon-Fri
	if dayFilter == "all" || dayFilter == "weekday" {
		return chunks
	}

	// For specific days (monday through sunday), filter to only include
	// chunks that match that day of week
	var targetWeekday time.Weekday
	switch dayFilter {
	case "monday":
		targetWeekday = time.Monday
	case "tuesday":
		targetWeekday = time.Tuesday
	case "wednesday":
		targetWeekday = time.Wednesday
	case "thursday":
		targetWeekday = time.Thursday
	case "friday":
		targetWeekday = time.Friday
	case "saturday":
		targetWeekday = time.Saturday
	case "sunday":
		targetWeekday = time.Sunday
	default:
		// Unknown filter, return all chunks
		return chunks
	}

	// Filter chunks to only include dates matching the target weekday
	var filtered []dateChunk
	for _, chunk := range chunks {
		if chunk.start.Weekday() == targetWeekday {
			filtered = append(filtered, chunk)
		}
	}

	return filtered
}

func convertDayFilter(filter string) *string {
	switch filter {
	case "weekday", "monday", "tuesday", "wednesday", "thursday", "friday":
		// All weekdays map to WEEKDAY
		// The filtering by specific day happens in filterChunksByDayOfWeek
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
	ctx context.Context,
	hspClient *external.HSPClient,
	hspResponses []*external.HSPServiceMetricsResponse,
	routeID int,
	query domain.RouteReliabilityQuery,
	startDate, endDate time.Time,
) (*domain.RouteReliabilityMetrics, error) {
	// Aggregate all services from multiple responses
	var allServices []external.HSPService
	for i, hspResp := range hspResponses {
		allServices = append(allServices, hspResp.Services...)
		if len(hspResp.Services) == 0 {
			// Log when we get empty responses to help debug
			fmt.Printf("WARNING: HSP response %d returned 0 services\n", i)
		}
	}

	fmt.Printf("Total services aggregated from %d responses: %d\n", len(hspResponses), len(allServices))

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
		}, nil
	}

	// Collect all RIDs and check for cancellations
	var allRIDs []string
	for _, service := range allServices {
		allRIDs = append(allRIDs, service.ServiceAttributesMetrics.RIDs...)
	}

	fmt.Printf("Checking %d trains for cancellations...\n", len(allRIDs))

	// Get cancellation count by checking serviceDetails for each RID
	cancellationCount, err := getCancellationCount(ctx, hspClient, allRIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get cancellation data: %w", err)
	}

	fmt.Printf("Found %d cancelled trains out of %d total\n", cancellationCount, len(allRIDs))

	// Extract and aggregate metrics from all HSP services
	// HSP provides metrics with different tolerance values (e.g., 0, 5, 10, 15 minutes)
	var (
		totalTolerance0    int
		totalNotTolerance0 int
		totalTolerance5    int
		totalNotTolerance5 int
		totalTolerance15   int
		totalTolerance30   int
	)

	// First, scan all services to see what tolerance values are present
	toleranceValuesFound := make(map[string]bool)
	for _, service := range allServices {
		for _, metric := range service.Metrics {
			if metric.GlobalTolerance {
				toleranceValuesFound[metric.ToleranceValue] = true
			}
		}
	}
	fmt.Printf("Unique tolerance values found (global only): %v\n", toleranceValuesFound)

	// Aggregate metrics from all services
	for i, service := range allServices {
		if i < 3 {
			// Log first 3 services to debug
			fmt.Printf("Service %d has %d metrics:\n", i, len(service.Metrics))
			for j, metric := range service.Metrics {
				fmt.Printf("  Metric %d: tolerance_value=%s, global_tolerance=%v, num_tolerance=%s, num_not_tolerance=%s\n",
					j, metric.ToleranceValue, metric.GlobalTolerance, metric.NumTolerance, metric.NumNotTolerance)
			}
		}

		for _, metric := range service.Metrics {
			if metric.ToleranceValue == "0" && metric.GlobalTolerance {
				// tolerance_value=0 means perfectly on time (0 delay)
				// num_tolerance = trains within 0 min (perfectly on time)
				// num_not_tolerance = trains with ANY delay
				if numTol, err := strconv.Atoi(metric.NumTolerance); err == nil {
					totalTolerance0 += numTol
				}
				if numNotTol, err := strconv.Atoi(metric.NumNotTolerance); err == nil {
					totalNotTolerance0 += numNotTol
				}
			}
			if metric.ToleranceValue == "5" && metric.GlobalTolerance {
				// Parse string values to integers
				if numTol, err := strconv.Atoi(metric.NumTolerance); err == nil {
					totalTolerance5 += numTol
				}
				if numNotTol, err := strconv.Atoi(metric.NumNotTolerance); err == nil {
					totalNotTolerance5 += numNotTol
				}
			}
			if metric.ToleranceValue == "15" && metric.GlobalTolerance {
				if numTol, err := strconv.Atoi(metric.NumTolerance); err == nil {
					totalTolerance15 += numTol
				}
			}
			if metric.ToleranceValue == "30" && metric.GlobalTolerance {
				if numTol, err := strconv.Atoi(metric.NumTolerance); err == nil {
					totalTolerance30 += numTol
				}
			}
		}
	}

	fmt.Printf("After aggregating: totalTolerance0=%d, totalNotTolerance0=%d, totalTolerance5=%d, totalNotTolerance5=%d, totalTolerance15=%d, totalTolerance30=%d\n",
		totalTolerance0, totalNotTolerance0, totalTolerance5, totalNotTolerance5, totalTolerance15, totalTolerance30)

	// Calculate on-time rate from tolerance=0 data (perfectly on time vs any delay)
	// Note: We removed tolerance parameter support due to API timeouts
	var totalServices int
	var onTimeRate float64

	if totalTolerance0+totalNotTolerance0 > 0 {
		totalServices = totalTolerance0 + totalNotTolerance0
		onTimeRate = float64(totalTolerance0) / float64(totalServices) * 100
	}

	// Calculate cancellation rate from real data
	totalTrains := len(allRIDs)
	cancellationRate := 0.0
	if totalTrains > 0 {
		cancellationRate = float64(cancellationCount) / float64(totalTrains) * 100
	}

	// Calculate reliability score (simplified without delay distribution)
	// Score = (OnTime% × 0.6) + ((100 - Cancellation%) × 0.4)
	// Simplified since we don't have accurate delay distribution data
	reliabilityScore := (onTimeRate * 0.6) + ((100 - cancellationRate) * 0.4)
	if reliabilityScore < 0 {
		reliabilityScore = 0
	}
	if reliabilityScore > 100 {
		reliabilityScore = 100
	}

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
		// Note: Delay distribution not available without tolerance data
		// Frontend should not display delay distribution
		AvgDelayMinutes:    0,
		MedianDelayMinutes: 0,
		P95DelayMinutes:    0,
		P99DelayMinutes:    0,
		Pct0To5MinLate:     0,
		Pct5To15MinLate:    0,
		Pct15To30MinLate:   0,
		Pct30PlusMinLate:   0,
		ReliabilityScore:   reliabilityScore,
		ComputedAt:         time.Now(),
	}

	return metrics, nil
}

func getCancellationCount(ctx context.Context, hspClient *external.HSPClient, rids []string) (int, error) {
	if len(rids) == 0 {
		return 0, nil
	}

	// Use worker pool to check RIDs in parallel (limit concurrency to avoid overwhelming API)
	const maxConcurrency = 5
	semaphore := make(chan struct{}, maxConcurrency)

	type result struct {
		cancelled bool
		err       error
	}

	results := make(chan result, len(rids))

	// Launch goroutines for each RID
	for _, rid := range rids {
		go func(r string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			req := external.HSPServiceDetailsRequest{RID: r}
			details, err := hspClient.GetServiceDetails(ctx, req)
			if err != nil {
				results <- result{cancelled: false, err: err}
				return
			}

			// Check if any location was cancelled
			cancelled := false
			for _, loc := range details.Locations {
				if loc.LateCanc {
					cancelled = true
					break
				}
			}

			results <- result{cancelled: cancelled, err: nil}
		}(rid)
	}

	// Collect results
	cancelCount := 0
	for i := 0; i < len(rids); i++ {
		res := <-results
		if res.err != nil {
			// Log but don't fail - cancellation data is nice-to-have
			fmt.Printf("Error getting service details: %v\n", res.err)
			continue
		}
		if res.cancelled {
			cancelCount++
		}
	}

	return cancelCount, nil
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
