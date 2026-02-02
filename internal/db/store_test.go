package db_test

import (
	"path/filepath"
	"testing"

	"claude-schedule/internal/db"

	"github.com/stretchr/testify/require"
)

func openTestStore(t *testing.T) *db.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func TestOpenCreatesDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "subdir", "test.db")
	store, err := db.Open(dbPath)
	require.NoError(t, err)
	require.NotNil(t, store)
	store.Close()
}

func TestOpenCreatesJobsTable(t *testing.T) {
	store := openTestStore(t)
	jobs, err := store.GetJobs()
	require.NoError(t, err)
	require.Empty(t, jobs)
}
