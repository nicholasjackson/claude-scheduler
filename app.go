package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"claude-schedule/internal/db"
	"claude-schedule/internal/executor"
	"claude-schedule/internal/scheduler"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// App struct
type App struct {
	store    *db.Store
	sched    *scheduler.Scheduler
	notifier *notifications.NotificationService
}

// NewApp creates a new App application struct
func NewApp(store *db.Store, notifier *notifications.NotificationService) *App {
	return &App{store: store, notifier: notifier}
}

// ServiceStartup is called when the app starts via the Wails v3 service lifecycle.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	emit := func(eventName string, data ...interface{}) {
		app := application.Get()
		app.Event.Emit(eventName, data...)
	}
	a.sched = scheduler.New(a.store, emit, executor.ClaudeExecute, 60*time.Second)
	a.sched.SetNotifyFunc(a.sendNotification)
	a.sched.Start(ctx)
	return nil
}

// sendNotification fires a native OS notification for job status changes.
// It recovers from panics because the notification backend (e.g. dbus on Linux)
// may have a nil connection on environments like WSL2.
func (a *App) sendNotification(jobName string, status string) {
	if a.notifier == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("notification: recovered from panic: %v", r)
		}
	}()

	var title, body string
	switch status {
	case "running":
		title = "Job Started"
		body = jobName + " is now running"
	case "success":
		title = "Job Completed"
		body = jobName + " finished successfully"
	case "failed":
		title = "Job Failed"
		body = jobName + " failed"
	case "waiting":
		title = "Job Needs Input"
		body = jobName + " is waiting for your answer"
	default:
		return
	}
	_ = a.notifier.SendNotification(notifications.NotificationOptions{
		ID:    "job-" + jobName + "-" + status,
		Title: title,
		Body:  body,
	})
}

// ServiceShutdown is called when the app is closing.
func (a *App) ServiceShutdown() error {
	if a.sched != nil {
		a.sched.Stop()
	}
	if a.store != nil {
		a.store.Close()
	}
	return nil
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetJobs returns all scheduled jobs.
func (a *App) GetJobs() ([]db.Job, error) {
	return a.store.GetJobs()
}

// GetJob returns a single job by ID.
func (a *App) GetJob(id string) (db.Job, error) {
	return a.store.GetJob(id)
}

// CreateJob inserts a new job and returns it with the generated ID.
func (a *App) CreateJob(job db.Job) (db.Job, error) {
	return a.store.CreateJob(job)
}

// UpdateJob updates an existing job.
func (a *App) UpdateJob(job db.Job) (db.Job, error) {
	return a.store.UpdateJob(job)
}

// DeleteJob removes a job by ID.
func (a *App) DeleteJob(id string) error {
	return a.store.DeleteJob(id)
}

// GetRunsForJob returns the recent run history for a job.
func (a *App) GetRunsForJob(jobID string) ([]db.JobRun, error) {
	return a.store.GetRunsForJob(jobID)
}

// GetMCPServers returns all configured MCP servers.
func (a *App) GetMCPServers() ([]db.MCPServer, error) {
	return a.store.GetMCPServers()
}

// CreateMCPServer adds a new MCP server configuration.
func (a *App) CreateMCPServer(srv db.MCPServer) (db.MCPServer, error) {
	return a.store.CreateMCPServer(srv)
}

// UpdateMCPServer updates an existing MCP server configuration.
func (a *App) UpdateMCPServer(srv db.MCPServer) (db.MCPServer, error) {
	return a.store.UpdateMCPServer(srv)
}

// DeleteMCPServer removes an MCP server by ID.
func (a *App) DeleteMCPServer(id string) error {
	return a.store.DeleteMCPServer(id)
}

// GetMCPServersForJob returns the MCP servers associated with a job.
func (a *App) GetMCPServersForJob(jobID string) ([]db.MCPServer, error) {
	return a.store.GetMCPServersForJob(jobID)
}

// SetJobMCPServers replaces all MCP server associations for a job.
func (a *App) SetJobMCPServers(jobID string, serverIDs []string) error {
	return a.store.SetJobMCPServers(jobID, serverIDs)
}

// RunJobNow triggers immediate execution of a job.
func (a *App) RunJobNow(jobID string) error {
	return a.sched.RunNow(jobID)
}

// AnswerQuestion sends the user's answer to a waiting job and resumes execution.
func (a *App) AnswerQuestion(jobID string, answer string) error {
	return a.sched.AnswerQuestion(jobID, answer)
}
