package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VeyronSakai/gh-runner-log/internal/domain/entity"
	testhelpers "github.com/VeyronSakai/gh-runner-log/test"
)

func TestFetchRunnerJobHistoryFiltersAndOrdersJobs(t *testing.T) {
	runner := &entity.Runner{ID: 42, Name: "runner-1"}

	start1 := time.Date(2025, 11, 16, 12, 0, 0, 0, time.UTC)
	end1 := start1.Add(5 * time.Minute)
	start2 := start1.Add(-time.Hour)
	jobs := []*entity.Job{
		{ID: 1, RunnerID: ptrInt64(99), StartedAt: &start1, CompletedAt: &end1},
		{ID: 2, RunnerID: ptrInt64(42), StartedAt: &start2},
		{ID: 3, RunnerID: ptrInt64(42), StartedAt: &start1},
	}

	runnerLogger := NewRunnerLogger(&testhelpers.StubJobRepository{Jobs: jobs}, &testhelpers.StubRunnerRepository{Runner: runner})
	history, err := runnerLogger.FetchRunnerJobHistory(context.Background(), "runner-1", 5)
	if err != nil {
		t.Fatalf("FetchRunnerJobHistory error: %v", err)
	}

	if len(history.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(history.Jobs))
	}
	if history.Jobs[0].ID != 3 {
		t.Fatalf("expected job 3 first, got %d", history.Jobs[0].ID)
	}
	if history.Jobs[1].ID != 2 {
		t.Fatalf("expected job 2 second, got %d", history.Jobs[1].ID)
	}
}

func TestFetchRunnerJobHistory_PrunesByLimit(t *testing.T) {
	runner := &entity.Runner{ID: 7, Name: "runner"}
	jobs := make([]*entity.Job, 0, 20)
	for i := 0; i < 20; i++ {
		started := time.Date(2025, 11, 16, 0, i, 0, 0, time.UTC)
		jobs = append(jobs, &entity.Job{ID: int64(i), RunnerID: ptrInt64(7), StartedAt: &started})
	}

	runnerLogger := NewRunnerLogger(&testhelpers.StubJobRepository{Jobs: jobs}, &testhelpers.StubRunnerRepository{Runner: runner})
	history, err := runnerLogger.FetchRunnerJobHistory(context.Background(), "runner", 5)
	if err != nil {
		t.Fatalf("FetchRunnerJobHistory error: %v", err)
	}

	if len(history.Jobs) != 5 {
		t.Fatalf("expected limit 5, got %d", len(history.Jobs))
	}
	if history.Jobs[0].ID != 4 {
		t.Fatalf("expected newest job id 4, got %d", history.Jobs[0].ID)
	}
}

func TestFetchRunnerJobHistory_PropagatesRunnerError(t *testing.T) {
	expected := errors.New("runner missing")
	runnerLogger := NewRunnerLogger(&testhelpers.StubJobRepository{}, &testhelpers.FailingRunnerRepository{Err: expected})

	_, err := runnerLogger.FetchRunnerJobHistory(context.Background(), "runner", 1)
	if !errors.Is(err, expected) {
		t.Fatalf("expected runner error, got %v", err)
	}
}

func TestFetchRunnerJobHistory_PropagatesJobError(t *testing.T) {
	runner := &entity.Runner{ID: 1, Name: "runner"}
	expected := errors.New("jobs fail")
	runnerLogger := NewRunnerLogger(&testhelpers.StubJobRepository{Err: expected}, &testhelpers.StubRunnerRepository{Runner: runner})

	_, err := runnerLogger.FetchRunnerJobHistory(context.Background(), "runner", 1)
	if !errors.Is(err, expected) {
		t.Fatalf("expected job error, got %v", err)
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
