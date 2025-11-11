// package net provides common network configuration constants and utilities
// for the skyeye application.
//
// The constants defined in this package represent recommended defaults for
// network operations. While they are exported and can be referenced, they
// are primarily intended as centralized configuration values used throughout
// the codebase. Operators needing different timeout behavior should configure
// their systems accordingly (e.g., adjusting infrastructure-level timeouts,
// network quality, or load balancers).
package net

import "time"

const (
	// DefaultConnectionTimeout is the default timeout for establishing network connections.
	// This applies to initial connection attempts for TCP, UDP resolution, and similar operations.
	DefaultConnectionTimeout = 30 * time.Second

	// ReadTimeoutMultiplier defines how much longer read timeouts should be compared to
	// connection timeouts for streaming data. A multiplier of 2 allows for transient
	// network delays while still detecting truly dead connections.
	//
	// Rationale: A 2x multiplier provides a balance between responsiveness to network
	// issues and tolerance for normal network latency/congestion. Lower values may cause
	// false positives during temporary network slowdowns; higher values delay detection
	// of truly dead connections.
	ReadTimeoutMultiplier = 2

	// DeadlineRefreshDivisor determines how frequently read deadlines should be refreshed
	// for long-lived streaming connections. The refresh interval is calculated as:
	// readTimeout / DeadlineRefreshDivisor. A divisor of 2 means we refresh at the
	// halfway point, ensuring the deadline never expires during active streaming.
	//
	// Rationale: Refreshing at the halfway point (divisor of 2) ensures the deadline is
	// always fresh with minimal overhead. This prevents legitimate long-running streams
	// from timing out while still allowing the full timeout period to detect dead connections.
	DeadlineRefreshDivisor = 2

	// ReconnectDelay is the delay before retrying a failed connection attempt.
	// This applies to automatic reconnection logic after connection failures.
	ReconnectDelay = 10 * time.Second

	// GRPCKeepaliveTime is the interval at which gRPC keepalive pings are sent.
	// This matches the default connection timeout to detect dead connections quickly.
	GRPCKeepaliveTime = 30 * time.Second

	// GRPCKeepaliveTimeout is how long to wait for a keepalive ping acknowledgment
	// before considering the connection dead.
	GRPCKeepaliveTimeout = 10 * time.Second

	// OpenAIHTTPTimeout is the timeout for HTTP requests to the OpenAI API.
	// Audio transcription can be slow for large files, so this is set higher
	// than typical HTTP timeouts.
	OpenAIHTTPTimeout = 60 * time.Second
)

// CalculateReadTimeout returns the recommended read timeout based on a connection timeout.
// This is useful for streaming connections where reads may take longer than initial connection.
func CalculateReadTimeout(connectionTimeout time.Duration) time.Duration {
	return connectionTimeout * ReadTimeoutMultiplier
}

// CalculateDeadlineRefreshInterval returns the interval at which read deadlines should
// be refreshed for long-lived streaming connections.
func CalculateDeadlineRefreshInterval(readTimeout time.Duration) time.Duration {
	return readTimeout / DeadlineRefreshDivisor
}
