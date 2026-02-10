import { ScheduledJob } from "./types";

export const mockJobs: ScheduledJob[] = [
  {
    id: "1",
    name: "Database Backup",
    startDate: "2026-01-31T02:00",
    intervalValue: 1,
    intervalUnit: "days",
    prompt: "Back up the production database to S3",
    active: true,
    nextRun: "2026-02-01T02:00:00Z",
    lastRun: "2026-01-31T02:00:00Z",
    status: "success",
    output: `[2026-01-31 02:00:01] Starting database backup...
[2026-01-31 02:00:03] Connecting to PostgreSQL at localhost:5432
[2026-01-31 02:00:03] Dumping database "app_production"...
[2026-01-31 02:00:15] Backup written to /backups/app_production_20260131.sql.gz (24.3 MB)
[2026-01-31 02:00:16] Uploading to S3 bucket "backups-prod"...
[2026-01-31 02:00:22] Upload complete.
[2026-01-31 02:00:22] Backup completed successfully.`,
    pendingQuestion: "",
  },
  {
    id: "2",
    name: "Log Rotation",
    startDate: "2026-01-26T00:00",
    intervalValue: 1,
    intervalUnit: "weeks",
    prompt: "Rotate application log files",
    active: true,
    nextRun: "2026-02-02T00:00:00Z",
    lastRun: "2026-01-26T00:00:00Z",
    status: "failed",
    output: `[2026-01-26 00:00:01] Starting log rotation...
[2026-01-26 00:00:02] Rotating /var/log/app/access.log
[2026-01-26 00:00:02] ERROR: Permission denied: /var/log/app/access.log
[2026-01-26 00:00:02] Log rotation failed with exit code 1.`,
    pendingQuestion: "",
  },
  {
    id: "3",
    name: "Health Check",
    startDate: "2026-01-31T14:00",
    intervalValue: 5,
    intervalUnit: "minutes",
    prompt: "Run health checks against all API endpoints",
    active: true,
    nextRun: "2026-01-31T14:05:00Z",
    lastRun: "2026-01-31T14:00:00Z",
    status: "running",
    output: `[2026-01-31 14:00:00] Running health checks...
[2026-01-31 14:00:01] Checking API endpoint: https://api.example.com/health
[2026-01-31 14:00:01] API: OK (response 200, 45ms)
[2026-01-31 14:00:02] Checking database connectivity...`,
    pendingQuestion: "",
  },
  {
    id: "4",
    name: "Report Generation",
    startDate: "2026-01-27T08:00",
    intervalValue: 1,
    intervalUnit: "weeks",
    prompt: "Generate weekly analytics report and email to the team",
    active: true,
    nextRun: "2026-02-02T08:00:00Z",
    lastRun: "2026-01-27T08:00:00Z",
    status: "success",
    output: `[2026-01-27 08:00:01] Generating weekly report...
[2026-01-27 08:00:05] Querying analytics data for 2026-01-20 to 2026-01-26
[2026-01-27 08:00:12] Report generated: /reports/weekly_20260127.pdf
[2026-01-27 08:00:13] Emailing report to team@example.com
[2026-01-27 08:00:14] Done.`,
    pendingQuestion: "",
  },
  {
    id: "5",
    name: "Cache Warmup",
    startDate: "2026-01-31T06:00",
    intervalValue: 1,
    intervalUnit: "days",
    prompt: "Warm up the application cache after nightly maintenance",
    active: false,
    nextRun: "2026-02-01T06:00:00Z",
    lastRun: "2026-01-31T06:00:00Z",
    status: "pending",
    output: "No output yet. Job has not run.",
    pendingQuestion: "",
  },
];
