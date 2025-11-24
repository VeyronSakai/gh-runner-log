package cmd

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		expectError bool
	}{
		{
			name:        "standard Go duration - hours",
			input:       "24h",
			expected:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "standard Go duration - minutes",
			input:       "30m",
			expected:    30 * time.Minute,
			expectError: false,
		},
		{
			name:        "standard Go duration - seconds",
			input:       "45s",
			expected:    45 * time.Second,
			expectError: false,
		},
		{
			name:        "days - single day",
			input:       "1d",
			expected:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "days - multiple days",
			input:       "7d",
			expected:    7 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "days - 30 days",
			input:       "30d",
			expected:    30 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "weeks - single week",
			input:       "1w",
			expected:    7 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "weeks - multiple weeks",
			input:       "2w",
			expected:    2 * 7 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid unit",
			input:       "5x",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("parseDuration(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseDuration(%q) unexpected error: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}

func TestParseSince(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		validate    func(t *testing.T, result time.Time)
	}{
		{
			name:        "empty string defaults to 24h",
			input:       "",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				// Result should be approximately 24 hours ago from now
				// We can't check exact time since parseSince uses time.Now()
				// Just verify it's not zero
				if result.IsZero() {
					t.Error("expected non-zero time for empty string")
				}
			},
		},
		{
			name:        "duration - 24 hours",
			input:       "24h",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				if result.IsZero() {
					t.Error("expected non-zero time")
				}
			},
		},
		{
			name:        "duration - 7 days",
			input:       "7d",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				if result.IsZero() {
					t.Error("expected non-zero time")
				}
			},
		},
		{
			name:        "duration - 2 weeks",
			input:       "2w",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				if result.IsZero() {
					t.Error("expected non-zero time")
				}
			},
		},
		{
			name:        "RFC3339 format",
			input:       "2025-11-17T10:00:00Z",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Date(2025, 11, 17, 10, 0, 0, 0, time.UTC)
				if !result.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:        "date format YYYY-MM-DD",
			input:       "2025-11-17",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC)
				if !result.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:        "date format - different date",
			input:       "2025-01-01",
			expectError: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
				if !result.Equal(expected) {
					t.Errorf("expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:        "invalid format",
			input:       "invalid-date",
			expectError: true,
			validate:    nil,
		},
		{
			name:        "invalid duration",
			input:       "5x",
			expectError: true,
			validate:    nil,
		},
		{
			name:        "partial date",
			input:       "2025-11",
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSince(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("parseSince(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseSince(%q) unexpected error: %v", tt.input, err)
				}
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestParseSince_DurationRelativeToNow(t *testing.T) {
	// Test that duration-based parsing is relative to current time
	before := time.Now()
	result, err := parseSince("1h")
	after := time.Now()
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Result should be approximately 1 hour before now
	// Account for test execution time by checking a range
	expectedMin := before.Add(-1*time.Hour - 1*time.Second)
	expectedMax := after.Add(-1 * time.Hour + 1*time.Second)
	
	if result.Before(expectedMin) || result.After(expectedMax) {
		t.Errorf("parseSince('1h') = %v, expected between %v and %v", result, expectedMin, expectedMax)
	}
}
