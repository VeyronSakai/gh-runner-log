package debug

import (
	"context"
	"testing"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
)

func TestJobRepositoryImpl_FetchJobHistory_TimeFiltering(t *testing.T) {
	// Create test data with different timestamps
	now := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	oneHourAgo := now.Add(-1 * time.Hour)
	oneDayAgo := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)
	oneWeekAgo := now.Add(-7 * 24 * time.Hour)

	runnerID := int64(123)
	
	jobs := []*entity.Job{
		{ID: 1, RunnerID: &runnerID, StartedAt: &now},           // Most recent
		{ID: 2, RunnerID: &runnerID, StartedAt: &oneHourAgo},    // 1 hour ago
		{ID: 3, RunnerID: &runnerID, StartedAt: &oneDayAgo},     // 1 day ago
		{ID: 4, RunnerID: &runnerID, StartedAt: &twoDaysAgo},    // 2 days ago
		{ID: 5, RunnerID: &runnerID, StartedAt: &oneWeekAgo},    // 1 week ago
	}

	tests := []struct {
		name         string
		createdAfter time.Time
		expectedIDs  []int64
	}{
		{
			name:         "no time filter",
			createdAfter: time.Time{}, // Zero time means no filter
			expectedIDs:  []int64{1, 2, 3, 4, 5},
		},
		{
			name:         "filter last 2 hours",
			createdAfter: now.Add(-2 * time.Hour),
			expectedIDs:  []int64{1, 2},
		},
		{
			name:         "filter last 24 hours",
			createdAfter: now.Add(-24 * time.Hour),
			expectedIDs:  []int64{1, 2, 3},
		},
		{
			name:         "filter last 3 days",
			createdAfter: now.Add(-3 * 24 * time.Hour),
			expectedIDs:  []int64{1, 2, 3, 4},
		},
		{
			name:         "filter last week",
			createdAfter: now.Add(-7 * 24 * time.Hour),
			expectedIDs:  []int64{1, 2, 3, 4, 5},
		},
		{
			name:         "filter very recent - no matches",
			createdAfter: now.Add(1 * time.Hour), // Future time
			expectedIDs:  []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &dataset{jobs: jobs}
			repo := &JobRepositoryImpl{
				ds:           ds,
				scope:        "",
				createdAfter: tt.createdAfter,
			}

			result, err := repo.FetchJobHistory(context.Background(), runnerID, 100)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expectedIDs) {
				t.Errorf("expected %d jobs, got %d", len(tt.expectedIDs), len(result))
			}

			// Check that we got the expected job IDs
			for i, expectedID := range tt.expectedIDs {
				if i >= len(result) {
					t.Errorf("missing expected job ID %d at index %d", expectedID, i)
					continue
				}
				if result[i].ID != expectedID {
					t.Errorf("at index %d: expected job ID %d, got %d", i, expectedID, result[i].ID)
				}
			}
		})
	}
}

func TestJobRepositoryImpl_FetchJobHistory_ScopeFiltering(t *testing.T) {
	runnerID := int64(123)
	startTime := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	
	jobs := []*entity.Job{
		{ID: 1, RunnerID: &runnerID, Repository: "acme-corp/frontend-app", StartedAt: &startTime},
		{ID: 2, RunnerID: &runnerID, Repository: "acme-corp/backend-app", StartedAt: &startTime},
		{ID: 3, RunnerID: &runnerID, Repository: "other-org/some-app", StartedAt: &startTime},
	}

	tests := []struct {
		name        string
		scope       string
		expectedIDs []int64
	}{
		{
			name:        "no scope filter",
			scope:       "",
			expectedIDs: []int64{1, 2, 3},
		},
		{
			name:        "filter by org",
			scope:       "acme-corp",
			expectedIDs: []int64{1, 2},
		},
		{
			name:        "filter by specific repo",
			scope:       "acme-corp/frontend-app",
			expectedIDs: []int64{1},
		},
		{
			name:        "filter by different org",
			scope:       "other-org",
			expectedIDs: []int64{3},
		},
		{
			name:        "no matches",
			scope:       "non-existent-org",
			expectedIDs: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &dataset{jobs: jobs}
			repo := &JobRepositoryImpl{
				ds:           ds,
				scope:        tt.scope,
				createdAfter: time.Time{}, // No time filter
			}

			result, err := repo.FetchJobHistory(context.Background(), runnerID, 100)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expectedIDs) {
				t.Errorf("expected %d jobs, got %d", len(tt.expectedIDs), len(result))
			}

			for i, expectedID := range tt.expectedIDs {
				if i >= len(result) {
					t.Errorf("missing expected job ID %d at index %d", expectedID, i)
					continue
				}
				if result[i].ID != expectedID {
					t.Errorf("at index %d: expected job ID %d, got %d", i, expectedID, result[i].ID)
				}
			}
		})
	}
}

func TestJobRepositoryImpl_FetchJobHistory_RunnerFiltering(t *testing.T) {
	runner1 := int64(123)
	runner2 := int64(456)
	startTime := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	
	jobs := []*entity.Job{
		{ID: 1, RunnerID: &runner1, StartedAt: &startTime},
		{ID: 2, RunnerID: &runner2, StartedAt: &startTime},
		{ID: 3, RunnerID: &runner1, StartedAt: &startTime},
		{ID: 4, RunnerID: nil, StartedAt: &startTime}, // No runner assigned
	}

	tests := []struct {
		name        string
		runnerID    int64
		expectedIDs []int64
	}{
		{
			name:        "filter by runner 123",
			runnerID:    123,
			expectedIDs: []int64{1, 3},
		},
		{
			name:        "filter by runner 456",
			runnerID:    456,
			expectedIDs: []int64{2},
		},
		{
			name:        "no runner filter (runnerID = 0)",
			runnerID:    0,
			expectedIDs: []int64{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &dataset{jobs: jobs}
			repo := &JobRepositoryImpl{
				ds:           ds,
				scope:        "",
				createdAfter: time.Time{},
			}

			result, err := repo.FetchJobHistory(context.Background(), tt.runnerID, 100)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expectedIDs) {
				t.Errorf("expected %d jobs, got %d", len(tt.expectedIDs), len(result))
			}

			for i, expectedID := range tt.expectedIDs {
				if i >= len(result) {
					t.Errorf("missing expected job ID %d at index %d", expectedID, i)
					continue
				}
				if result[i].ID != expectedID {
					t.Errorf("at index %d: expected job ID %d, got %d", i, expectedID, result[i].ID)
				}
			}
		})
	}
}

func TestJobRepositoryImpl_FetchJobHistory_CombinedFilters(t *testing.T) {
	// Test combining time filter, scope filter, and runner filter
	now := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	oneDayAgo := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)
	
	runner1 := int64(123)
	runner2 := int64(456)
	
	jobs := []*entity.Job{
		{ID: 1, RunnerID: &runner1, Repository: "acme-corp/app1", StartedAt: &now},
		{ID: 2, RunnerID: &runner1, Repository: "acme-corp/app2", StartedAt: &oneDayAgo},
		{ID: 3, RunnerID: &runner2, Repository: "acme-corp/app1", StartedAt: &now},
		{ID: 4, RunnerID: &runner1, Repository: "other-org/app1", StartedAt: &now},
		{ID: 5, RunnerID: &runner1, Repository: "acme-corp/app1", StartedAt: &twoDaysAgo},
	}

	ds := &dataset{jobs: jobs}
	repo := &JobRepositoryImpl{
		ds:           ds,
		scope:        "acme-corp",
		createdAfter: now.Add(-30 * time.Hour), // Last 30 hours
	}

	// Should match jobs: 1, 2 (runner1, acme-corp, within 30 hours)
	result, err := repo.FetchJobHistory(context.Background(), runner1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedIDs := []int64{1, 2}
	if len(result) != len(expectedIDs) {
		t.Errorf("expected %d jobs, got %d", len(expectedIDs), len(result))
	}

	for i, expectedID := range expectedIDs {
		if i >= len(result) {
			t.Errorf("missing expected job ID %d at index %d", expectedID, i)
			continue
		}
		if result[i].ID != expectedID {
			t.Errorf("at index %d: expected job ID %d, got %d", i, expectedID, result[i].ID)
		}
	}
}

func TestJobRepositoryImpl_FetchJobHistory_NilStartedAt(t *testing.T) {
	// Test that jobs with nil StartedAt are handled correctly with time filter
	now := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	oneDayAgo := now.Add(-24 * time.Hour)
	
	runnerID := int64(123)
	
	jobs := []*entity.Job{
		{ID: 1, RunnerID: &runnerID, StartedAt: &now},
		{ID: 2, RunnerID: &runnerID, StartedAt: nil}, // No start time
		{ID: 3, RunnerID: &runnerID, StartedAt: &oneDayAgo},
	}

	ds := &dataset{jobs: jobs}
	repo := &JobRepositoryImpl{
		ds:           ds,
		scope:        "",
		createdAfter: now.Add(-12 * time.Hour), // Last 12 hours
	}

	result, err := repo.FetchJobHistory(context.Background(), runnerID, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only match job 1 (job 2 has nil StartedAt and is excluded, job 3 is too old)
	expectedIDs := []int64{1}
	if len(result) != len(expectedIDs) {
		t.Errorf("expected %d jobs, got %d", len(expectedIDs), len(result))
	}

	if len(result) > 0 && result[0].ID != expectedIDs[0] {
		t.Errorf("expected job ID %d, got %d", expectedIDs[0], result[0].ID)
	}
}

func TestJobRepositoryImpl_FetchJobHistory_ZeroLimit(t *testing.T) {
	runnerID := int64(123)
	startTime := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	
	jobs := []*entity.Job{
		{ID: 1, RunnerID: &runnerID, StartedAt: &startTime},
	}

	ds := &dataset{jobs: jobs}
	repo := &JobRepositoryImpl{
		ds:           ds,
		scope:        "",
		createdAfter: time.Time{},
	}

	result, err := repo.FetchJobHistory(context.Background(), runnerID, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 jobs with limit 0, got %d", len(result))
	}
}
