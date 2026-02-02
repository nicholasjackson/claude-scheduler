package db

import (
	"fmt"

	"github.com/google/uuid"
)

// Valid interval units for job scheduling.
var validIntervalUnits = map[string]bool{
	"minutes": true,
	"hours":   true,
	"days":    true,
	"weeks":   true,
}

// Job represents a scheduled job persisted in the database.
type Job struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	StartDate     string `json:"startDate"`
	IntervalValue int    `json:"intervalValue"`
	IntervalUnit  string `json:"intervalUnit"`
	Prompt        string `json:"prompt"`
	Active        bool   `json:"active"`
	NextRun       string `json:"nextRun"`
	LastRun       string `json:"lastRun"`
	Status        string `json:"status"`
	Output        string `json:"output"`
}

func validateJob(j Job) error {
	if j.IntervalValue <= 0 {
		return fmt.Errorf("interval value must be greater than 0")
	}
	if !validIntervalUnits[j.IntervalUnit] {
		return fmt.Errorf("invalid interval unit: %s", j.IntervalUnit)
	}
	return nil
}

// GetJobs returns all jobs sorted by name.
func (s *Store) GetJobs() ([]Job, error) {
	rows, err := s.db.Query(
		"SELECT id, name, start_date, interval_value, interval_unit, prompt, active, next_run, last_run, status, output FROM jobs ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []Job{}
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Name, &j.StartDate, &j.IntervalValue, &j.IntervalUnit,
			&j.Prompt, &j.Active, &j.NextRun, &j.LastRun, &j.Status, &j.Output); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// GetJob returns a single job by ID.
func (s *Store) GetJob(id string) (Job, error) {
	var j Job
	err := s.db.QueryRow(
		"SELECT id, name, start_date, interval_value, interval_unit, prompt, active, next_run, last_run, status, output FROM jobs WHERE id = ?",
		id,
	).Scan(&j.ID, &j.Name, &j.StartDate, &j.IntervalValue, &j.IntervalUnit,
		&j.Prompt, &j.Active, &j.NextRun, &j.LastRun, &j.Status, &j.Output)
	return j, err
}

// CreateJob inserts a new job. It assigns a UUID if ID is empty and defaults status to "pending".
func (s *Store) CreateJob(j Job) (Job, error) {
	if err := validateJob(j); err != nil {
		return j, err
	}
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	if j.Status == "" {
		j.Status = "pending"
	}
	_, err := s.db.Exec(
		`INSERT INTO jobs (id, name, start_date, interval_value, interval_unit, prompt, active, next_run, last_run, status, output)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		j.ID, j.Name, j.StartDate, j.IntervalValue, j.IntervalUnit,
		j.Prompt, j.Active, j.NextRun, j.LastRun, j.Status, j.Output,
	)
	return j, err
}

// UpdateJob updates an existing job. Returns an error if the job does not exist.
func (s *Store) UpdateJob(j Job) (Job, error) {
	if err := validateJob(j); err != nil {
		return j, err
	}
	result, err := s.db.Exec(
		`UPDATE jobs SET name=?, start_date=?, interval_value=?, interval_unit=?, prompt=?, active=?, next_run=?, last_run=?, status=?, output=?
		 WHERE id=?`,
		j.Name, j.StartDate, j.IntervalValue, j.IntervalUnit,
		j.Prompt, j.Active, j.NextRun, j.LastRun, j.Status, j.Output, j.ID,
	)
	if err != nil {
		return j, err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return j, fmt.Errorf("job not found: %s", j.ID)
	}
	return j, nil
}

// ResetRunningJobs resets any jobs stuck in "running" status back to "failed".
// This handles the case where the app crashed or was killed mid-execution.
func (s *Store) ResetRunningJobs() (int64, error) {
	result, err := s.db.Exec(
		`UPDATE jobs SET status='failed', output='interrupted: app was restarted' WHERE status='running'`,
	)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return n, nil
}

// DeleteJob removes a job by ID. Returns an error if the job does not exist.
func (s *Store) DeleteJob(id string) error {
	result, err := s.db.Exec("DELETE FROM jobs WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("job not found: %s", id)
	}
	return nil
}
