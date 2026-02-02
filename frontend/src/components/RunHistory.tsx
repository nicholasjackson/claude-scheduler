import { useEffect, useMemo, useState } from "react";
import { marked } from "marked";
import { GetRunsForJob, OnEvent } from "../wailsbridge";
import { JobRun } from "../types";
import { formatTime } from "../utils";

interface Props {
  jobId: string;
}

const statusDot: Record<string, string> = {
  success: "bg-green-400",
  failed: "bg-red-400",
  running: "bg-yellow-400 animate-pulse",
};

function duration(start: string, end: string): string {
  if (!start || !end) return "—";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  if (isNaN(ms) || ms < 0) return "—";
  const secs = Math.floor(ms / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const remSecs = secs % 60;
  if (mins < 60) return `${mins}m ${remSecs}s`;
  const hrs = Math.floor(mins / 60);
  const remMins = mins % 60;
  return `${hrs}h ${remMins}m`;
}

function RunOutput({ output }: { output: string }) {
  const html = useMemo(() => {
    if (!output) return "";
    return marked.parse(output, { async: false }) as string;
  }, [output]);

  if (!output) {
    return (
      <div className="border-t border-gray-700 px-3 py-2">
        <p className="text-xs text-gray-500 italic">(no output)</p>
      </div>
    );
  }

  return (
    <div
      className="border-t border-gray-700 px-3 py-2 max-h-80 overflow-y-auto text-sm text-gray-300"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

export default function RunHistory({ jobId }: Props) {
  const [runs, setRuns] = useState<JobRun[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const fetchRuns = () => {
    GetRunsForJob(jobId).then((data) => {
      const list = data ?? [];
      setRuns(list);
      // Auto-expand latest run if nothing is expanded yet.
      if (list.length > 0 && expandedId === null) {
        setExpandedId(list[0].id);
      }
    });
  };

  useEffect(() => {
    fetchRuns();
    const off = OnEvent("jobs:updated", fetchRuns);
    return off;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [jobId]);

  if (runs.length === 0) {
    return (
      <p className="text-sm text-gray-500 italic">No runs yet.</p>
    );
  }

  return (
    <div className="space-y-2">
      {runs.map((run) => {
        const isExpanded = expandedId === run.id;
        const dot = statusDot[run.status] ?? "bg-gray-400";
        return (
          <div key={run.id} className="border border-gray-700 rounded">
            <button
              onClick={() => setExpandedId(isExpanded ? null : run.id)}
              className="w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-gray-800 transition-colors"
            >
              <span className={`w-2 h-2 rounded-full shrink-0 ${dot}`} />
              <span className="text-xs text-gray-300 flex-1">
                {formatTime(run.startedAt)}
              </span>
              <span className="text-xs text-gray-500">
                {run.status === "running" ? "running…" : duration(run.startedAt, run.endedAt)}
              </span>
              <span className="text-xs text-gray-500">
                {isExpanded ? "▲" : "▼"}
              </span>
            </button>
            {isExpanded && (
              <RunOutput output={run.output} />
            )}
          </div>
        );
      })}
    </div>
  );
}
