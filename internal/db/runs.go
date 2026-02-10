package db

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	maxRunsPerJob   = 10
	maxOutputBytes  = 100 * 1024 // 100 KB
	truncatedMarker = "\n\n[truncated]"
)

// JobRun represents a single execution of a scheduled job.
type JobRun struct {
	ID              string `json:"id"`
	JobID           string `json:"jobId"`
	StartedAt       string `json:"startedAt"`
	EndedAt         string `json:"endedAt"`
	Status          string `json:"status"`
	Output          string `json:"output"`
	PendingQuestion string `json:"pendingQuestion"`
}

// truncateOutput trims output to maxOutputBytes and appends a marker if truncated.
func truncateOutput(s string) string {
	if len(s) <= maxOutputBytes {
		return s
	}
	return s[:maxOutputBytes-len(truncatedMarker)] + truncatedMarker
}

// CreateRun inserts a new job run with an auto-generated UUID.
func (s *Store) CreateRun(run JobRun) (JobRun, error) {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	run.Output = truncateOutput(run.Output)

	_, err := s.db.Exec(
		`INSERT INTO job_runs (id, job_id, started_at, ended_at, status, output, pending_question)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.JobID, run.StartedAt, run.EndedAt, run.Status, run.Output, run.PendingQuestion,
	)
	return run, err
}

// UpdateRun updates an existing run's status, output, and ended_at.
func (s *Store) UpdateRun(run JobRun) error {
	run.Output = truncateOutput(run.Output)

	result, err := s.db.Exec(
		`UPDATE job_runs SET status=?, output=?, ended_at=?, pending_question=? WHERE id=?`,
		run.Status, run.Output, run.EndedAt, run.PendingQuestion, run.ID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("run not found: %s", run.ID)
	}
	return nil
}

// GetRunsForJob returns the most recent runs for a job, ordered newest first.
func (s *Store) GetRunsForJob(jobID string) ([]JobRun, error) {
	rows, err := s.db.Query(
		`SELECT id, job_id, started_at, ended_at, status, output, pending_question
		 FROM job_runs WHERE job_id = ?
		 ORDER BY started_at DESC LIMIT ?`,
		jobID, maxRunsPerJob,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []JobRun{}
	for rows.Next() {
		var r JobRun
		if err := rows.Scan(&r.ID, &r.JobID, &r.StartedAt, &r.EndedAt, &r.Status, &r.Output, &r.PendingQuestion); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// GetLatestRun returns the most recent run for a job.
func (s *Store) GetLatestRun(jobID string) (JobRun, error) {
	var r JobRun
	err := s.db.QueryRow(
		`SELECT id, job_id, started_at, ended_at, status, output, pending_question
		 FROM job_runs WHERE job_id = ?
		 ORDER BY started_at DESC LIMIT 1`,
		jobID,
	).Scan(&r.ID, &r.JobID, &r.StartedAt, &r.EndedAt, &r.Status, &r.Output, &r.PendingQuestion)
	return r, err
}

// PruneRuns deletes all but the most recent maxRunsPerJob runs for a job.
func (s *Store) PruneRuns(jobID string) error {
	_, err := s.db.Exec(
		`DELETE FROM job_runs WHERE job_id = ? AND id NOT IN (
			SELECT id FROM job_runs WHERE job_id = ?
			ORDER BY started_at DESC LIMIT ?
		)`,
		jobID, jobID, maxRunsPerJob,
	)
	return err
}

// DeleteRunsForJob removes all runs for a given job.
func (s *Store) DeleteRunsForJob(jobID string) error {
	_, err := s.db.Exec(`DELETE FROM job_runs WHERE job_id = ?`, jobID)
	return err
}
