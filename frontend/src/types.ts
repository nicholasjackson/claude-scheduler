export type JobStatus = "success" | "failed" | "running" | "pending";

export type IntervalUnit = "minutes" | "hours" | "days" | "weeks";

export interface JobRun {
  id: string;
  jobId: string;
  startedAt: string;
  endedAt: string;
  status: JobStatus;
  output: string;
}

export type MCPServerType = "http" | "stdio";

export interface MCPServer {
  id: string;
  name: string;
  type: MCPServerType;
  url: string;
  command: string;
  args: string;
  env: string;
  headers: string;
}

export interface ScheduledJob {
  id: string;
  name: string;
  startDate: string;
  intervalValue: number;
  intervalUnit: IntervalUnit;
  prompt: string;
  active: boolean;
  nextRun: string;
  lastRun: string;
  status: JobStatus;
  output: string;
}
