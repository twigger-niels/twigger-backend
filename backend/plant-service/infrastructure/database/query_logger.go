package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SlowQueryThreshold defines the duration after which a query is considered slow
const SlowQueryThreshold = 100 * time.Millisecond

// QueryLogger wraps database operations with performance logging
type QueryLogger struct {
	db *sql.DB
}

// NewQueryLogger creates a new query logger wrapper
func NewQueryLogger(db *sql.DB) *QueryLogger {
	return &QueryLogger{db: db}
}

// LoggedQueryRowContext executes a query and logs if it's slow
func (ql *QueryLogger) LoggedQueryRowContext(ctx context.Context, query string, args ...interface{}) (*sql.Row, time.Duration) {
	start := time.Now()
	row := ql.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > SlowQueryThreshold {
		logSlowQuery("QueryRow", query, args, duration)
	}

	return row, duration
}

// LoggedQueryContext executes a query and logs if it's slow
func (ql *QueryLogger) LoggedQueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, time.Duration, error) {
	start := time.Now()
	rows, err := ql.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > SlowQueryThreshold {
		logSlowQuery("Query", query, args, duration)
	}

	return rows, duration, err
}

// LoggedExecContext executes a command and logs if it's slow
func (ql *QueryLogger) LoggedExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, time.Duration, error) {
	start := time.Now()
	result, err := ql.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > SlowQueryThreshold {
		logSlowQuery("Exec", query, args, duration)
	}

	return result, duration, err
}

// LogQueryDuration logs query performance with context
func LogQueryDuration(operation, entityType string, duration time.Duration, recordCount int) {
	if duration > SlowQueryThreshold {
		log.Printf("[SLOW QUERY] operation=%s entity=%s duration=%dms records=%d",
			operation, entityType, duration.Milliseconds(), recordCount)
	}
}

// LogQueryDurationWithFilter logs query performance with filter details
func LogQueryDurationWithFilter(operation, entityType string, filter map[string]interface{}, duration time.Duration, recordCount int) {
	if duration > SlowQueryThreshold {
		log.Printf("[SLOW QUERY] operation=%s entity=%s filter=%v duration=%dms records=%d",
			operation, entityType, filter, duration.Milliseconds(), recordCount)
	}
}

// Helper function to log slow queries
func logSlowQuery(queryType, query string, args []interface{}, duration time.Duration) {
	// Sanitize query for logging (remove excess whitespace)
	sanitizedQuery := sanitizeQuery(query)

	log.Printf("[SLOW QUERY] type=%s duration=%dms query=%s args=%v",
		queryType, duration.Milliseconds(), sanitizedQuery, sanitizeArgs(args))
}

// sanitizeQuery removes excess whitespace from queries for cleaner logs
func sanitizeQuery(query string) string {
	// Simple sanitization - replace multiple spaces/newlines with single space
	result := ""
	prevSpace := false
	for _, char := range query {
		if char == ' ' || char == '\n' || char == '\t' || char == '\r' {
			if !prevSpace {
				result += " "
				prevSpace = true
			}
		} else {
			result += string(char)
			prevSpace = false
		}
	}

	// Truncate very long queries
	if len(result) > 500 {
		return result[:497] + "..."
	}
	return result
}

// sanitizeArgs removes sensitive data from logged arguments
func sanitizeArgs(args []interface{}) []interface{} {
	sanitized := make([]interface{}, len(args))
	for i, arg := range args {
		// Don't log very long strings (might be GeoJSON)
		if str, ok := arg.(string); ok && len(str) > 100 {
			sanitized[i] = fmt.Sprintf("<string len=%d>", len(str))
		} else {
			sanitized[i] = arg
		}
	}
	return sanitized
}

// WithQueryLogging is a helper to wrap repository operations with logging
func WithQueryLogging(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	if duration > SlowQueryThreshold {
		if err != nil {
			log.Printf("[SLOW QUERY] operation=%s duration=%dms status=error error=%v",
				operation, duration.Milliseconds(), err)
		} else {
			log.Printf("[SLOW QUERY] operation=%s duration=%dms status=success",
				operation, duration.Milliseconds())
		}
	}

	return err
}

// QueryMetrics tracks query performance metrics
type QueryMetrics struct {
	TotalQueries   int64
	SlowQueries    int64
	TotalDuration  time.Duration
	SlowestQuery   time.Duration
	SlowestOp      string
}

var metrics QueryMetrics

// RecordQueryMetric records a query for metrics tracking
func RecordQueryMetric(operation string, duration time.Duration) {
	metrics.TotalQueries++
	metrics.TotalDuration += duration

	if duration > SlowQueryThreshold {
		metrics.SlowQueries++
	}

	if duration > metrics.SlowestQuery {
		metrics.SlowestQuery = duration
		metrics.SlowestOp = operation
	}
}

// GetQueryMetrics returns current query metrics
func GetQueryMetrics() QueryMetrics {
	return metrics
}

// ResetQueryMetrics resets all metrics
func ResetQueryMetrics() {
	metrics = QueryMetrics{}
}
