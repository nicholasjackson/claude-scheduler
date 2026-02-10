package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"claude-schedule/internal/db"
	"claude-schedule/internal/executor"
)

// EmitFunc is the signature for a Wails-style event emitter.
type EmitFunc func(eventName string, data ...interface{})

// NotifyFunc is called when a job changes status (e.g. "running", "success", "failed").
type NotifyFunc func(jobName string, status string)

// ExecuteFunc defines how a job is executed. It receives the job and its
// associated MCP servers, and returns an ExecuteResult and an error.
type ExecuteFunc func(ctx context.Context, job db.Job, mcpServers []db.MCPServer) (executor.ExecuteResult, error)

// AnswerFunc defines how a question answer is sent back to Claude.
type AnswerFunc func(ctx context.Context, job db.Job, mcpServers []db.MCPServer, answer string) (executor.ExecuteResult, error)

// Scheduler polls the database at a fixed interval and runs due jobs sequentially.
type Scheduler struct {
	store    *db.Store
	emitFn   EmitFunc
	notifyFn NotifyFunc
	execFn   ExecuteFunc
	answerFn AnswerFunc
	interval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a Scheduler. Pass the tick interval and an execution function.
// If execFn is nil a default mock executor (30 s sleep) is used.
func New(store *db.Store, emitFn EmitFunc, execFn ExecuteFunc, interval time.Duration) *Scheduler {
	if execFn == nil {
		execFn = mockExecute
	}
	return &Scheduler{
		store:    store,
		emitFn:   emitFn,
		execFn:   execFn,
		answerFn: executor.ClaudeAnswer,
		interval: interval,
	}
}

// SetNotifyFunc sets an optional callback for job status change notifications.
func (s *Scheduler) SetNotifyFunc(fn NotifyFunc) {
	s.notifyFn = fn
}

// Start begins the background tick loop. It is safe to call only once.
// It resets any jobs left in "running" state from a previous crash.
func (s *Scheduler) Start(parent context.Context) {
	if n, err := s.store.ResetRunningJobs(); err != nil {
		log.Printf("scheduler: failed to reset running jobs: %v", err)
	} else if n > 0 {
		log.Printf("scheduler: reset %d stale running job(s)", n)
		s.emit()
	}

	s.ctx, s.cancel = context.WithCancel(parent)
	s.wg.Add(1)
	go s.loop()
}

// Stop cancels the tick loop and waits for in-flight work to finish.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) loop() {
	defer s.wg.Done()

	// Run one tick immediately on startup.
	s.tick()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Scheduler) tick() {
	jobs, err := s.store.GetJobs()
	if err != nil {
		log.Printf("scheduler: failed to load jobs: %v", err)
		return
	}

	now := time.Now().UTC()
	for i := range jobs {
		// Check for cancellation between jobs.
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		if isDue(jobs[i], now) {
			s.executeJob(&jobs[i], now)
		}
	}
}

// isDue returns true when a job should be executed right now.
func isDue(job db.Job, now time.Time) bool {
	if !job.Active {
		return false
	}
	if job.Status == "running" || job.Status == "waiting" {
		return false
	}

	interval := intervalDuration(job.IntervalValue, job.IntervalUnit)
	if interval == 0 {
		return false
	}

	// Use LastRun if available, otherwise fall back to StartDate.
	ref := job.LastRun
	if ref == "" {
		ref = job.StartDate
	}
	if ref == "" {
		return false
	}

	refTime, err := parseTime(ref)
	if err != nil {
		log.Printf("scheduler: cannot parse reference time %q for job %s: %v", ref, job.ID, err)
		return false
	}

	return !refTime.Add(interval).After(now)
}

// finishExecution processes the result of a CLI invocation, detecting questions
// and updating job/run state accordingly.
func (s *Scheduler) finishExecution(job *db.Job, run *db.JobRun, result executor.ExecuteResult, execErr error) {
	if execErr != nil {
		job.Status = "failed"
		job.Output = execErr.Error()
		job.PendingQuestion = ""
	} else {
		// Check for a pending question in the raw output.
		question := executor.DetectQuestion(result.RawLines)
		if question != "" {
			job.Status = "waiting"
			job.Output = result.Transcript
			job.PendingQuestion = question
		} else {
			job.Status = "success"
			job.Output = result.Transcript
			job.PendingQuestion = ""
		}
	}

	if _, err := s.store.UpdateJob(*job); err != nil {
		log.Printf("scheduler: failed to update job %s after execution: %v", job.ID, err)
	}

	// Update the run record.
	if run != nil && run.ID != "" {
		run.Status = job.Status
		run.Output = job.Output
		run.PendingQuestion = job.PendingQuestion
		if job.Status != "waiting" {
			run.EndedAt = time.Now().UTC().Format(time.RFC3339)
		}
		if err := s.store.UpdateRun(*run); err != nil {
			log.Printf("scheduler: failed to update run %s: %v", run.ID, err)
		}
		if err := s.store.PruneRuns(job.ID); err != nil {
			log.Printf("scheduler: failed to prune runs for job %s: %v", job.ID, err)
		}
	}

	s.emit()
	s.notify(job.Name, job.Status)
}

func (s *Scheduler) executeJob(job *db.Job, now time.Time) {
	// Mark as running.
	job.Status = "running"
	job.Output = ""
	job.PendingQuestion = ""
	if _, err := s.store.UpdateJob(*job); err != nil {
		log.Printf("scheduler: failed to mark job %s running: %v", job.ID, err)
		return
	}
	s.emit()
	s.notify(job.Name, "running")

	// Create a run record.
	run, err := s.store.CreateRun(db.JobRun{
		JobID:     job.ID,
		StartedAt: now.Format(time.RFC3339),
		Status:    "running",
	})
	if err != nil {
		log.Printf("scheduler: failed to create run for job %s: %v", job.ID, err)
	}

	// Fetch MCP servers for this job.
	mcpServers, err := s.store.GetMCPServersForJob(job.ID)
	if err != nil {
		log.Printf("scheduler: failed to load MCP servers for job %s: %v", job.ID, err)
	}

	// Execute.
	result, execErr := s.execFn(s.ctx, *job, mcpServers)

	// Update timing fields.
	job.LastRun = now.Format(time.RFC3339)
	interval := intervalDuration(job.IntervalValue, job.IntervalUnit)
	job.NextRun = now.Add(interval).Format(time.RFC3339)

	s.finishExecution(job, &run, result, execErr)
}

func (s *Scheduler) emit() {
	if s.emitFn != nil {
		s.emitFn("jobs:updated")
	}
}

func (s *Scheduler) notify(jobName string, status string) {
	if s.notifyFn != nil {
		s.notifyFn(jobName, status)
	}
}

// RunNow triggers immediate execution of the given job in the background.
// Returns an error if the job is already running.
func (s *Scheduler) RunNow(jobID string) error {
	job, err := s.store.GetJob(jobID)
	if err != nil {
		return err
	}
	if job.Status == "running" {
		return fmt.Errorf("job is already running")
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.executeJob(&job, time.Now().UTC())
	}()

	return nil
}

// AnswerQuestion sends the user's answer to a waiting job and resumes execution.
func (s *Scheduler) AnswerQuestion(jobID string, answer string) error {
	job, err := s.store.GetJob(jobID)
	if err != nil {
		return err
	}
	if job.Status != "waiting" {
		return fmt.Errorf("job is not waiting for an answer")
	}

	// Mark as running again.
	job.Status = "running"
	job.PendingQuestion = ""
	if _, err := s.store.UpdateJob(job); err != nil {
		return fmt.Errorf("updating job status: %w", err)
	}
	s.emit()

	// Get the latest run to append output to.
	run, err := s.store.GetLatestRun(jobID)
	if err != nil {
		log.Printf("scheduler: failed to get latest run for job %s: %v", jobID, err)
	}
	if run.ID != "" {
		run.Status = "running"
		run.PendingQuestion = ""
		if err := s.store.UpdateRun(run); err != nil {
			log.Printf("scheduler: failed to update run %s: %v", run.ID, err)
		}
	}
	s.emit()

	// Fetch MCP servers.
	mcpServers, err := s.store.GetMCPServersForJob(jobID)
	if err != nil {
		log.Printf("scheduler: failed to load MCP servers for job %s: %v", jobID, err)
	}

	// Resume the conversation with the answer.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		result, execErr := s.answerFn(s.ctx, job, mcpServers, answer)

		// Append new output to the existing run output.
		if run.ID != "" && execErr == nil {
			result.Transcript = run.Output + "\n\n" + result.Transcript
		}

		s.finishExecution(&job, &run, result, execErr)
	}()

	return nil
}

// intervalDuration converts the stored interval value+unit to a time.Duration.
func intervalDuration(value int, unit string) time.Duration {
	switch unit {
	case "minutes":
		return time.Duration(value) * time.Minute
	case "hours":
		return time.Duration(value) * time.Hour
	case "days":
		return time.Duration(value) * 24 * time.Hour
	case "weeks":
		return time.Duration(value) * 7 * 24 * time.Hour
	default:
		return 0
	}
}

// parseTime tries RFC3339 first, then the datetime-local format used by the frontend.
func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02T15:04", s)
}

// mockExecute is the default executor: sleeps for 30 seconds.
func mockExecute(ctx context.Context, _ db.Job, _ []db.MCPServer) (executor.ExecuteResult, error) {
	select {
	case <-time.After(30 * time.Second):
		return executor.ExecuteResult{Transcript: "Mock execution completed successfully."}, nil
	case <-ctx.Done():
		return executor.ExecuteResult{}, ctx.Err()
	}
}
