package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"claude-schedule/internal/db"
	"claude-schedule/internal/executor"

	"github.com/stretchr/testify/require"
)

func tempStore(t *testing.T) *db.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := db.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close(); os.Remove(path) })
	return store
}

func pastTime(d time.Duration) string {
	return time.Now().UTC().Add(-d).Format(time.RFC3339)
}

func futureTime(d time.Duration) string {
	return time.Now().UTC().Add(d).Format(time.RFC3339)
}

func createJob(t *testing.T, store *db.Store, name string, active bool, intervalValue int, intervalUnit string, lastRun string) db.Job {
	t.Helper()
	j := db.Job{
		Name:          name,
		StartDate:     "2026-01-01T00:00",
		IntervalValue: intervalValue,
		IntervalUnit:  intervalUnit,
		Active:        active,
		Status:        "pending",
		LastRun:       lastRun,
	}
	created, err := store.CreateJob(j)
	require.NoError(t, err)
	return created
}

// fastExec returns an executor that completes instantly.
func fastExec() ExecuteFunc {
	return func(_ context.Context, _ db.Job, _ []db.MCPServer) (executor.ExecuteResult, error) {
		return executor.ExecuteResult{Transcript: "done"}, nil
	}
}

func noopEmit(string, ...interface{}) {}

func TestIsDue(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		job      db.Job
		expected bool
	}{
		{
			name: "inactive job is not due",
			job: db.Job{
				Active:        false,
				IntervalValue: 1,
				IntervalUnit:  "minutes",
				LastRun:       pastTime(10 * time.Minute),
			},
			expected: false,
		},
		{
			name: "running job is not due",
			job: db.Job{
				Active:        true,
				Status:        "running",
				IntervalValue: 1,
				IntervalUnit:  "minutes",
				LastRun:       pastTime(10 * time.Minute),
			},
			expected: false,
		},
		{
			name: "job with last run in the past beyond interval is due",
			job: db.Job{
				Active:        true,
				Status:        "success",
				IntervalValue: 5,
				IntervalUnit:  "minutes",
				LastRun:       pastTime(10 * time.Minute),
			},
			expected: true,
		},
		{
			name: "job with last run recent is not due",
			job: db.Job{
				Active:        true,
				Status:        "success",
				IntervalValue: 1,
				IntervalUnit:  "hours",
				LastRun:       pastTime(5 * time.Minute),
			},
			expected: false,
		},
		{
			name: "never-run job uses start date",
			job: db.Job{
				Active:        true,
				Status:        "pending",
				IntervalValue: 1,
				IntervalUnit:  "minutes",
				StartDate:     pastTime(5 * time.Minute),
			},
			expected: true,
		},
		{
			name: "never-run job with future start date is not due",
			job: db.Job{
				Active:        true,
				Status:        "pending",
				IntervalValue: 1,
				IntervalUnit:  "minutes",
				StartDate:     futureTime(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "invalid interval unit is not due",
			job: db.Job{
				Active:        true,
				Status:        "pending",
				IntervalValue: 1,
				IntervalUnit:  "fortnights",
				LastRun:       pastTime(10 * time.Minute),
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isDue(tc.job, now)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSchedulerRunsDueJob(t *testing.T) {
	store := tempStore(t)
	job := createJob(t, store, "due-job", true, 1, "minutes", pastTime(10*time.Minute))

	emitted := 0
	emit := func(string, ...interface{}) { emitted++ }

	sched := New(store, emit, fastExec(), 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	sched.Start(ctx)

	// Wait for the initial tick to complete.
	time.Sleep(200 * time.Millisecond)
	cancel()
	sched.Stop()

	updated, err := store.GetJob(job.ID)
	require.NoError(t, err)
	require.Equal(t, "success", updated.Status)
	require.NotEmpty(t, updated.LastRun)
	require.NotEmpty(t, updated.NextRun)
	require.Equal(t, "done", updated.Output)
	// At least 2 emits: one for "running", one for "success"
	require.GreaterOrEqual(t, emitted, 2)
}

func TestSchedulerSkipsInactiveJob(t *testing.T) {
	store := tempStore(t)
	job := createJob(t, store, "inactive-job", false, 1, "minutes", pastTime(10*time.Minute))

	sched := New(store, noopEmit, fastExec(), 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	sched.Start(ctx)

	time.Sleep(200 * time.Millisecond)
	cancel()
	sched.Stop()

	updated, err := store.GetJob(job.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", updated.Status)
}

func TestSchedulerSkipsNotYetDueJob(t *testing.T) {
	store := tempStore(t)
	job := createJob(t, store, "not-due", true, 1, "hours", pastTime(5*time.Minute))

	sched := New(store, noopEmit, fastExec(), 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	sched.Start(ctx)

	time.Sleep(200 * time.Millisecond)
	cancel()
	sched.Stop()

	updated, err := store.GetJob(job.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", updated.Status)
}

func TestSchedulerStopsCleanly(t *testing.T) {
	store := tempStore(t)

	sched := New(store, noopEmit, fastExec(), 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	sched.Start(ctx)

	time.Sleep(100 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		sched.Stop()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return within 2 seconds")
	}
}

func TestIntervalDuration(t *testing.T) {
	require.Equal(t, 5*time.Minute, intervalDuration(5, "minutes"))
	require.Equal(t, 2*time.Hour, intervalDuration(2, "hours"))
	require.Equal(t, 3*24*time.Hour, intervalDuration(3, "days"))
	require.Equal(t, 1*7*24*time.Hour, intervalDuration(1, "weeks"))
	require.Equal(t, time.Duration(0), intervalDuration(1, "invalid"))
}

func TestParseTime(t *testing.T) {
	// RFC3339
	_, err := parseTime("2026-01-31T14:00:00Z")
	require.NoError(t, err)

	// datetime-local format
	_, err = parseTime("2026-01-31T14:00")
	require.NoError(t, err)

	// invalid
	_, err = parseTime("not-a-date")
	require.Error(t, err)
}
