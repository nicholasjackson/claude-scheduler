package db_test

import (
	"testing"

	"claude-schedule/internal/db"

	"github.com/stretchr/testify/require"
)

func validJob(name string) db.Job {
	return db.Job{
		Name:          name,
		StartDate:     "2026-02-01T00:00",
		IntervalValue: 15,
		IntervalUnit:  "minutes",
	}
}

func TestCreateJobAssignsID(t *testing.T) {
	store := openTestStore(t)
	job, err := store.CreateJob(validJob("Test"))
	require.NoError(t, err)
	require.NotEmpty(t, job.ID)
}

func TestCreateJobDefaultsPendingStatus(t *testing.T) {
	store := openTestStore(t)
	job, err := store.CreateJob(validJob("Test"))
	require.NoError(t, err)
	require.Equal(t, "pending", job.Status)
}

func TestGetJobReturnsCreatedJob(t *testing.T) {
	store := openTestStore(t)
	j := validJob("Backup")
	j.StartDate = "2026-02-01T02:00"
	j.IntervalValue = 1
	j.IntervalUnit = "days"
	created, err := store.CreateJob(j)
	require.NoError(t, err)

	fetched, err := store.GetJob(created.ID)
	require.NoError(t, err)
	require.Equal(t, "Backup", fetched.Name)
	require.Equal(t, "2026-02-01T02:00", fetched.StartDate)
	require.Equal(t, 1, fetched.IntervalValue)
	require.Equal(t, "days", fetched.IntervalUnit)
}

func TestGetJobsReturnsSortedByName(t *testing.T) {
	store := openTestStore(t)
	_, err := store.CreateJob(validJob("Zebra"))
	require.NoError(t, err)
	_, err = store.CreateJob(validJob("Alpha"))
	require.NoError(t, err)

	jobs, err := store.GetJobs()
	require.NoError(t, err)
	require.Len(t, jobs, 2)
	require.Equal(t, "Alpha", jobs[0].Name)
	require.Equal(t, "Zebra", jobs[1].Name)
}

func TestGetJobsReturnsEmptySlice(t *testing.T) {
	store := openTestStore(t)
	jobs, err := store.GetJobs()
	require.NoError(t, err)
	require.NotNil(t, jobs)
	require.Len(t, jobs, 0)
}

func TestUpdateJobModifiesFields(t *testing.T) {
	store := openTestStore(t)
	created, err := store.CreateJob(validJob("Old"))
	require.NoError(t, err)

	created.Name = "New"
	created.Status = "running"
	_, err = store.UpdateJob(created)
	require.NoError(t, err)

	fetched, err := store.GetJob(created.ID)
	require.NoError(t, err)
	require.Equal(t, "New", fetched.Name)
	require.Equal(t, "running", fetched.Status)
}

func TestUpdateJobReturnsErrorForMissingJob(t *testing.T) {
	store := openTestStore(t)
	j := validJob("Missing")
	j.ID = "nonexistent"
	_, err := store.UpdateJob(j)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestDeleteJobRemovesJob(t *testing.T) {
	store := openTestStore(t)
	created, err := store.CreateJob(validJob("ToDelete"))
	require.NoError(t, err)

	err = store.DeleteJob(created.ID)
	require.NoError(t, err)

	_, err = store.GetJob(created.ID)
	require.Error(t, err)
}

func TestDeleteJobReturnsErrorForMissingJob(t *testing.T) {
	store := openTestStore(t)
	err := store.DeleteJob("nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestGetJobReturnsErrorForMissingJob(t *testing.T) {
	store := openTestStore(t)
	_, err := store.GetJob("nonexistent")
	require.Error(t, err)
}

func TestCreateJobPersistsPromptAndActive(t *testing.T) {
	store := openTestStore(t)
	j := validJob("PromptTest")
	j.Prompt = "Summarize the news"
	j.Active = true
	job, err := store.CreateJob(j)
	require.NoError(t, err)

	fetched, err := store.GetJob(job.ID)
	require.NoError(t, err)
	require.Equal(t, "Summarize the news", fetched.Prompt)
	require.True(t, fetched.Active)
}

func TestCreateJobDefaultsActiveToFalse(t *testing.T) {
	store := openTestStore(t)
	job, err := store.CreateJob(validJob("Inactive"))
	require.NoError(t, err)

	fetched, err := store.GetJob(job.ID)
	require.NoError(t, err)
	require.False(t, fetched.Active)
}

func TestUpdateJobModifiesPromptAndActive(t *testing.T) {
	store := openTestStore(t)
	j := validJob("Original")
	j.Prompt = "Old prompt"
	j.Active = true
	created, err := store.CreateJob(j)
	require.NoError(t, err)

	created.Prompt = "New prompt"
	created.Active = false
	_, err = store.UpdateJob(created)
	require.NoError(t, err)

	fetched, err := store.GetJob(created.ID)
	require.NoError(t, err)
	require.Equal(t, "New prompt", fetched.Prompt)
	require.False(t, fetched.Active)
}

func TestCreateJobPersistsScheduleFields(t *testing.T) {
	store := openTestStore(t)
	job, err := store.CreateJob(db.Job{
		Name:          "ScheduleTest",
		StartDate:     "2026-03-15T09:30",
		IntervalValue: 2,
		IntervalUnit:  "weeks",
	})
	require.NoError(t, err)

	fetched, err := store.GetJob(job.ID)
	require.NoError(t, err)
	require.Equal(t, "2026-03-15T09:30", fetched.StartDate)
	require.Equal(t, 2, fetched.IntervalValue)
	require.Equal(t, "weeks", fetched.IntervalUnit)
}

func TestCreateJobValidatesIntervalValue(t *testing.T) {
	store := openTestStore(t)
	_, err := store.CreateJob(db.Job{
		Name:          "BadInterval",
		IntervalValue: 0,
		IntervalUnit:  "hours",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "interval value must be greater than 0")
}

func TestCreateJobValidatesIntervalUnit(t *testing.T) {
	store := openTestStore(t)
	_, err := store.CreateJob(db.Job{
		Name:          "BadUnit",
		IntervalValue: 1,
		IntervalUnit:  "fortnights",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid interval unit")
}

func TestUpdateJobModifiesScheduleFields(t *testing.T) {
	store := openTestStore(t)
	created, err := store.CreateJob(db.Job{
		Name:          "Original",
		StartDate:     "2026-01-01T00:00",
		IntervalValue: 1,
		IntervalUnit:  "hours",
	})
	require.NoError(t, err)

	created.StartDate = "2026-06-01T12:00"
	created.IntervalValue = 30
	created.IntervalUnit = "minutes"
	_, err = store.UpdateJob(created)
	require.NoError(t, err)

	fetched, err := store.GetJob(created.ID)
	require.NoError(t, err)
	require.Equal(t, "2026-06-01T12:00", fetched.StartDate)
	require.Equal(t, 30, fetched.IntervalValue)
	require.Equal(t, "minutes", fetched.IntervalUnit)
}
