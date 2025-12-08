package net

import (
	"testing"
	"time"
)

func TestCalculateReadTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		connectionTimeout time.Duration
		expected          time.Duration
	}{
		{
			name:              "default connection timeout",
			connectionTimeout: DefaultConnectionTimeout,
			expected:          60 * time.Second,
		},
		{
			name:              "10 second connection timeout",
			connectionTimeout: 10 * time.Second,
			expected:          20 * time.Second,
		},
		{
			name:              "1 minute connection timeout",
			connectionTimeout: 1 * time.Minute,
			expected:          2 * time.Minute,
		},
		{
			name:              "zero timeout",
			connectionTimeout: 0,
			expected:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := CalculateReadTimeout(tt.connectionTimeout)
			if actual != tt.expected {
				t.Errorf("CalculateReadTimeout(%v) = %v, expected %v", tt.connectionTimeout, actual, tt.expected)
			}
		})
	}
}

func TestCalculateDeadlineRefreshInterval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		readTimeout time.Duration
		expected    time.Duration
	}{
		{
			name:        "60 second read timeout",
			readTimeout: 60 * time.Second,
			expected:    30 * time.Second,
		},
		{
			name:        "20 second read timeout",
			readTimeout: 20 * time.Second,
			expected:    10 * time.Second,
		},
		{
			name:        "2 minute read timeout",
			readTimeout: 2 * time.Minute,
			expected:    1 * time.Minute,
		},
		{
			name:        "zero timeout",
			readTimeout: 0,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := CalculateDeadlineRefreshInterval(tt.readTimeout)
			if actual != tt.expected {
				t.Errorf("CalculateDeadlineRefreshInterval(%v) = %v, expected %v", tt.readTimeout, actual, tt.expected)
			}
		})
	}
}

func TestCalculateReadTimeoutMultiplier(t *testing.T) {
	t.Parallel()
	// Verify that the multiplier constant is correctly applied
	connectionTimeout := 15 * time.Second
	readTimeout := CalculateReadTimeout(connectionTimeout)

	expected := time.Duration(int64(connectionTimeout) * ReadTimeoutMultiplier)
	if readTimeout != expected {
		t.Errorf("ReadTimeoutMultiplier not correctly applied: got %v, expected %v", readTimeout, expected)
	}
}

func TestCalculateDeadlineRefreshDivisor(t *testing.T) {
	t.Parallel()
	// Verify that the divisor constant is correctly applied
	readTimeout := 40 * time.Second
	refreshInterval := CalculateDeadlineRefreshInterval(readTimeout)

	expected := time.Duration(int64(readTimeout) / DeadlineRefreshDivisor)
	if refreshInterval != expected {
		t.Errorf("DeadlineRefreshDivisor not correctly applied: got %v, expected %v", refreshInterval, expected)
	}
}

func TestEndToEndTimeoutCalculation(t *testing.T) {
	t.Parallel()
	// Test the complete flow from connection timeout to refresh interval
	connectionTimeout := DefaultConnectionTimeout // 30s

	readTimeout := CalculateReadTimeout(connectionTimeout)
	if readTimeout != 60*time.Second {
		t.Errorf("Expected read timeout of 60s for default connection timeout, got %v", readTimeout)
	}

	refreshInterval := CalculateDeadlineRefreshInterval(readTimeout)
	if refreshInterval != 30*time.Second {
		t.Errorf("Expected refresh interval of 30s for 60s read timeout, got %v", refreshInterval)
	}
}

func TestConstantValues(t *testing.T) {
	t.Parallel()
	// Verify the constant values match expected defaults
	if DefaultConnectionTimeout != 30*time.Second {
		t.Errorf("DefaultConnectionTimeout = %v, expected 30s", DefaultConnectionTimeout)
	}

	if ReadTimeoutMultiplier != 2 {
		t.Errorf("ReadTimeoutMultiplier = %v, expected 2", ReadTimeoutMultiplier)
	}

	if DeadlineRefreshDivisor != 2 {
		t.Errorf("DeadlineRefreshDivisor = %v, expected 2", DeadlineRefreshDivisor)
	}

	if ReconnectDelay != 10*time.Second {
		t.Errorf("ReconnectDelay = %v, expected 10s", ReconnectDelay)
	}

	if GRPCKeepaliveTime != 30*time.Second {
		t.Errorf("GRPCKeepaliveTime = %v, expected 30s", GRPCKeepaliveTime)
	}

	if GRPCKeepaliveTimeout != 10*time.Second {
		t.Errorf("GRPCKeepaliveTimeout = %v, expected 10s", GRPCKeepaliveTimeout)
	}

	if OpenAIHTTPTimeout != 60*time.Second {
		t.Errorf("OpenAIHTTPTimeout = %v, expected 60s", OpenAIHTTPTimeout)
	}
}
