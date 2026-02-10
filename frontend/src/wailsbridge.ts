import { Call, Events } from "@wailsio/runtime";
import type { ScheduledJob, JobRun, MCPServer } from "./types";

// Call Go service methods by name. These will be replaced by auto-generated
// bindings once `wails3 generate bindings` is run.

export function GetJobs(): Promise<ScheduledJob[]> {
  return Call.ByName("main.App.GetJobs");
}

export function CreateJob(job: ScheduledJob): Promise<ScheduledJob> {
  return Call.ByName("main.App.CreateJob", job);
}

export function UpdateJob(job: ScheduledJob): Promise<ScheduledJob> {
  return Call.ByName("main.App.UpdateJob", job);
}

export function DeleteJob(id: string): Promise<void> {
  return Call.ByName("main.App.DeleteJob", id);
}

export function GetRunsForJob(jobId: string): Promise<JobRun[]> {
  return Call.ByName("main.App.GetRunsForJob", jobId);
}

// MCP Server methods.

export function GetMCPServers(): Promise<MCPServer[]> {
  return Call.ByName("main.App.GetMCPServers");
}

export function CreateMCPServer(srv: MCPServer): Promise<MCPServer> {
  return Call.ByName("main.App.CreateMCPServer", srv);
}

export function UpdateMCPServer(srv: MCPServer): Promise<MCPServer> {
  return Call.ByName("main.App.UpdateMCPServer", srv);
}

export function DeleteMCPServer(id: string): Promise<void> {
  return Call.ByName("main.App.DeleteMCPServer", id);
}

export function GetMCPServersForJob(jobId: string): Promise<MCPServer[]> {
  return Call.ByName("main.App.GetMCPServersForJob", jobId);
}

export function SetJobMCPServers(jobId: string, serverIds: string[]): Promise<void> {
  return Call.ByName("main.App.SetJobMCPServers", jobId, serverIds);
}

export function RunJobNow(jobId: string): Promise<void> {
  return Call.ByName("main.App.RunJobNow", jobId);
}

export function AnswerQuestion(jobId: string, answer: string): Promise<void> {
  return Call.ByName("main.App.AnswerQuestion", jobId, answer);
}

// Event helpers wrapping the v3 Events API.
export function OnEvent(name: string, callback: (data: unknown) => void): () => void {
  return Events.On(name, callback);
}
