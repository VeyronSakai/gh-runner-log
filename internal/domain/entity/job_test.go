package entity

import (
	"testing"
	"time"
)

func TestJob_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "completed job",
			status:   "completed",
			expected: true,
		},
		{
			name:     "in_progress job",
			status:   "in_progress",
			expected: false,
		},
		{
			name:     "queued job",
			status:   "queued",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Status: tt.status}
			if got := job.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_IsAssignedToRunner(t *testing.T) {
	runnerID := int64(123)

	tests := []struct {
		name       string
		job        *Job
		runnerID   int64
		expected   bool
	}{
		{
			name:     "job assigned to runner",
			job:      &Job{RunnerID: &runnerID},
			runnerID: 123,
			expected: true,
		},
		{
			name:     "job assigned to different runner",
			job:      &Job{RunnerID: &runnerID},
			runnerID: 456,
			expected: false,
		},
		{
			name:     "job not assigned to any runner",
			job:      &Job{RunnerID: nil},
			runnerID: 123,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.job.IsAssignedToRunner(tt.runnerID); got != tt.expected {
				t.Errorf("IsAssignedToRunner() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_GetExecutionDuration(t *testing.T) {
	startTime := time.Date(2025, 11, 15, 10, 0, 0, 0, time.UTC)
	completedTime := time.Date(2025, 11, 15, 10, 5, 30, 0, time.UTC)
	expectedDuration := 5*time.Minute + 30*time.Second

	tests := []struct {
		name     string
		job      *Job
		expected time.Duration
	}{
		{
			name: "completed job with start and end times",
			job: &Job{
				StartedAt:   &startTime,
				CompletedAt: &completedTime,
			},
			expected: expectedDuration,
		},
		{
			name: "job without start time",
			job: &Job{
				StartedAt:   nil,
				CompletedAt: &completedTime,
			},
			expected: 0,
		},
		{
			name: "job without completion time",
			job: &Job{
				StartedAt:   &startTime,
				CompletedAt: nil,
			},
			expected: 0,
		},
		{
			name: "job without any times",
			job: &Job{
				StartedAt:   nil,
				CompletedAt: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.job.GetExecutionDuration(); got != tt.expected {
				t.Errorf("GetExecutionDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}
