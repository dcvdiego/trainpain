package external

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dcvdiego/trainpain-backend/internal/domain"
)

const (
	PineappleClassesURL = "https://www.pineapple.uk.com/pages/dance-classes-for-adults-at-pineapple-dance-studios"
)

// PineappleScraper handles scraping of Pineapple Dance Studios website
type PineappleScraper struct {
	timeout time.Duration
}

// NewPineappleScraper creates a new Pineapple scraper
func NewPineappleScraper() *PineappleScraper {
	return &PineappleScraper{
		timeout: 60 * time.Second,
	}
}

// ScrapeClasses scrapes all classes from the Pineapple Studios website
func (s *PineappleScraper) ScrapeClasses(ctx context.Context) ([]domain.ScrapedClassData, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Create browser options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	// Create browser context
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var htmlContent string

	// Navigate to the page and get HTML content
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(PineappleClassesURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for dynamic content to load
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scrape page: %w", err)
	}

	// Parse the HTML content
	classes, err := s.parseHTMLContent(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return classes, nil
}

// parseHTMLContent parses the HTML and extracts class data
// This is a placeholder - actual implementation will depend on the HTML structure
func (s *PineappleScraper) parseHTMLContent(html string) ([]domain.ScrapedClassData, error) {
	// TODO: This needs to be implemented based on actual HTML structure
	// Since we got 403, we need to inspect the page manually first
	// For now, returning a placeholder implementation

	classes := []domain.ScrapedClassData{}

	// Parse the HTML to find class information
	// This is a simplified example - actual parsing will be more complex

	// Example patterns to look for:
	// - Class names (e.g., "Ballet Basics", "Hip Hop Intermediate")
	// - Times (e.g., "18:30", "7:00 PM")
	// - Days (Monday, Tuesday, etc.)
	// - Instructors
	// - Available seats

	return classes, nil
}

// parseTime converts various time formats to HH:MM:SS format
func parseTime(timeStr string) (string, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Try parsing various formats
	formats := []string{
		"15:04",
		"3:04 PM",
		"3:04PM",
		"15:04:05",
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t.Format("15:04:05"), nil
		}
	}

	return "", fmt.Errorf("unable to parse time: %s", timeStr)
}

// parseDayOfWeek converts day name to integer (0=Sunday, 1=Monday, ..., 6=Saturday)
func parseDayOfWeek(day string) int {
	day = strings.ToLower(strings.TrimSpace(day))

	switch day {
	case "sunday", "sun":
		return 0
	case "monday", "mon":
		return 1
	case "tuesday", "tue", "tues":
		return 2
	case "wednesday", "wed":
		return 3
	case "thursday", "thu", "thur", "thurs":
		return 4
	case "friday", "fri":
		return 5
	case "saturday", "sat":
		return 6
	default:
		return -1
	}
}

// extractNumber extracts the first number from a string
func extractNumber(s string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}

	num, _ := strconv.Atoi(match)
	return num
}

// ScrapeClassAvailability scrapes seat availability for a specific class
func (s *PineappleScraper) ScrapeClassAvailability(ctx context.Context, className string, dayOfWeek int, startTime string) (*domain.ScrapedClassData, error) {
	// This would navigate to the specific class booking page and get availability
	// For now, this is a placeholder

	// TODO: Implement class-specific availability scraping
	return nil, fmt.Errorf("not implemented yet")
}
