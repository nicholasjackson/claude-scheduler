import { useEffect, useState, useCallback, useRef } from "react";
import { GetJobs, CreateJob, UpdateJob, SetJobMCPServers, RunJobNow, OnEvent } from "./wailsbridge";
import { ScheduledJob } from "./types";
import { useToasts } from "./hooks/useToasts";
import JobList from "./components/JobList";
import JobDetail from "./components/JobDetail";
import JobForm from "./components/JobForm";
import MCPSettings from "./components/MCPSettings";
import ToastContainer from "./components/Toast";

type ViewMode = "detail" | "new" | "edit" | "settings";

function App() {
  const [jobs, setJobs] = useState<ScheduledJob[]>([]);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>("detail");
  const [saveError, setSaveError] = useState<string | null>(null);
  const { toasts, addToast, removeToast } = useToasts();
  const prevJobsRef = useRef<Map<string, string>>(new Map());

  const refreshJobs = useCallback(async (selectId?: string) => {
    const result = await GetJobs();
    const loaded = (result ?? []) as unknown as ScheduledJob[];

    // Detect status transitions and fire toasts.
    const prev = prevJobsRef.current;
    for (const job of loaded) {
      const prevStatus = prev.get(job.id);
      if (prevStatus === undefined) continue; // new job, skip
      if (prevStatus === job.status) continue; // no change

      if (job.status === "running" && prevStatus !== "running") {
        addToast(`${job.name} started`, "info");
      } else if (job.status === "success" && prevStatus === "running") {
        addToast(`${job.name} completed`, "success");
      } else if (job.status === "failed" && prevStatus === "running") {
        addToast(`${job.name} failed`, "error");
      }
    }

    // Update previous state for next comparison.
    const next = new Map<string, string>();
    for (const job of loaded) {
      next.set(job.id, job.status);
    }
    prevJobsRef.current = next;

    setJobs(loaded);
    if (selectId) {
      setSelectedJobId(selectId);
    } else {
      setSelectedJobId((prev) => {
        if (prev && loaded.find((j) => j.id === prev)) return prev;
        return loaded.length > 0 ? loaded[0].id : null;
      });
    }
  }, [addToast]);

  useEffect(() => {
    refreshJobs();
    const cancel = OnEvent("jobs:updated", () => refreshJobs());
    return () => { cancel(); };
  }, [refreshJobs]);

  const selectedJob: ScheduledJob | null =
    jobs.find((j) => j.id === selectedJobId) ?? null;

  const handleSelectJob = (id: string) => {
    setSelectedJobId(id);
    setViewMode("detail");
  };

  const handleNewJob = () => {
    setSaveError(null);
    setViewMode("new");
    setSelectedJobId(null);
  };

  const handleEditJob = () => {
    setViewMode("edit");
  };

  const handleRunNow = async (jobId: string) => {
    try {
      await RunJobNow(jobId);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      addToast(message, "error");
    }
  };

  const handleSaveJob = async (job: ScheduledJob, mcpServerIds: string[]) => {
    setSaveError(null);
    try {
      let savedId: string;
      if (job.id) {
        const updated = await UpdateJob(job);
        savedId = updated.id;
      } else {
        const created = await CreateJob(job);
        savedId = created.id;
      }
      // Save MCP server associations.
      await SetJobMCPServers(savedId, mcpServerIds);
      await refreshJobs(savedId);
      setViewMode("detail");
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setSaveError(message);
      console.error("Failed to save job:", err);
    }
  };

  const handleCancelForm = () => {
    setSaveError(null);
    setViewMode("detail");
  };

  return (
    <div className="h-screen flex bg-gray-900 text-white font-sans">
      <div className="w-[30%] min-w-[240px] border-r border-gray-700 flex flex-col">
        <JobList
          jobs={jobs}
          selectedJobId={selectedJobId}
          onSelectJob={handleSelectJob}
          onNewJob={handleNewJob}
        />
        <div className="px-4 py-3 border-t border-gray-700">
          <button
            onClick={() => setViewMode("settings")}
            className={`w-full text-left text-xs font-medium px-2 py-1.5 rounded transition-colors ${
              viewMode === "settings"
                ? "text-blue-400 bg-gray-800"
                : "text-gray-500 hover:text-gray-300"
            }`}
          >
            MCP Servers
          </button>
        </div>
      </div>
      <div className="flex-1">
        {viewMode === "settings" && (
          <MCPSettings onClose={() => setViewMode("detail")} />
        )}
        {viewMode === "new" && (
          <JobForm job={null} onSave={handleSaveJob} onCancel={handleCancelForm} saveError={saveError} />
        )}
        {viewMode === "edit" && selectedJob && (
          <JobForm job={selectedJob} onSave={handleSaveJob} onCancel={handleCancelForm} saveError={saveError} />
        )}
        {viewMode === "detail" && (
          <JobDetail job={selectedJob} onEdit={handleEditJob} onRunNow={handleRunNow} />
        )}
      </div>
      <ToastContainer toasts={toasts} onDismiss={removeToast} />
    </div>
  );
}

export default App;
