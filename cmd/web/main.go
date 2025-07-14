package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ricardovhz/rinha-2025/internal"
)

func parseDateTime(dateStr string) time.Time {
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z07:00",
	}
	for _, f := range formats {
		t, err := time.Parse(f, dateStr)
		if err == nil {
			return t
		}
	}
	fmt.Printf("Failed to parse date %s with formats %v\n", dateStr, formats)
	return time.Time{} // Return zero value if no format matches
}

func main() {
	// Initialize a new Fiber app
	app := fiber.New()

	// Starts the backend selector
	defaultBackendHost := os.Getenv("DEFAULT_BACKEND_HOST")
	fallbackBackendHost := os.Getenv("FALLBACK_BACKEND_HOST")
	if defaultBackendHost == "" || fallbackBackendHost == "" {
		log.Fatal("Environment variables DEFAULT_BACKEND_HOST and FALLBACK_BACKEND_HOST must be set")
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	dbHost := os.Getenv("DB_HOST")
	summaryService := internal.NewSummaryService(dbHost)

	defaultBackend := internal.NewBackend(defaultBackendHost, &http.Client{})
	fallbackBackend := internal.NewBackend(fallbackBackendHost, &http.Client{})

	selector := internal.NewBackendSelector(ctx, defaultBackend, fallbackBackend, func(backend int, amount float64, requestedAt time.Time) {
		am := fmt.Sprintf("%.2f", amount)
		aa, _ := strconv.Atoi(strings.ReplaceAll(am, ".", ""))
		fmt.Printf("Payment processed by backend %d: amount=%.5f (%d) at %s\n", backend, amount, aa, requestedAt)
		summaryService.AddPayment(ctx, backend, int64(aa), requestedAt)
	})
	selector.Start()

	q := internal.NewQueue()
	w := internal.NewWorkersPool(selector, q, 10)
	defer w.Stop()

	// warm up
	for range 10 {
		summaryService.GetSummary(context.Background(), time.Now().Add(-1*time.Hour), time.Now())
	}

	app.Post("/payments", func(c fiber.Ctx) error {
		var req internal.PaymentsRequest
		err := c.Bind().JSON(&req)
		if err != nil {
			ctx.Err()
			return fmt.Errorf("failed to parse request: %w", err)
		}
		req.RequestedAt = time.Now().Format(time.RFC3339)
		q.Enqueue(req)
		c.Status(200)
		return nil
	})

	app.Get("/payments-summary", func(c fiber.Ctx) error {
		var query = make(map[string]string)
		err := c.Bind().Query(query)
		if err != nil {
			return fmt.Errorf("failed to parse query: %w", err)
		}
		from := query["from"]
		to := query["to"]
		fmt.Printf("Fetching payments summary from %s to %s\n", from, to)
		fromTime := parseDateTime(from)
		toTime := parseDateTime(to)
		time.Sleep(500 * time.Millisecond) // Simulate some processing delay

		sum, err := summaryService.GetSummary(c.Context(), fromTime, toTime)
		if err != nil {
			fmt.Printf("Failed to get summary: %v\n", err)
			return err
		}
		fmt.Printf("Summary from %s to %s: %+v %+v\n", from, to, sum.DefaultBackend, sum.FallbackBackend)
		c.JSON(map[string]any{
			"default": map[string]any{
				"totalRequests": sum.DefaultBackend.TotalRequests,
				"totalAmount":   float64(sum.DefaultBackend.TotalAmount) / 100.0,
			},
			"fallback": map[string]any{
				"totalRequests": sum.FallbackBackend.TotalRequests,
				"totalAmount":   float64(sum.FallbackBackend.TotalAmount) / 100.0,
			},
		})
		c.Status(200)
		return nil
	})

	// Start the server on port 9999
	log.Fatal(app.Listen(":9999"))
}
