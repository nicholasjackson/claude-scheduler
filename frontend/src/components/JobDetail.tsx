import { useMemo } from "react";
import { marked } from "marked";
import { ScheduledJob } from "../types";
import { formatInterval, formatTime } from "../utils";
import RunHistory from "./RunHistory";

interface Props {
  job: ScheduledJob | null;
  onEdit: () => void;
}

const statusLabels: Record<string, { text: string; color: string }> = {
  success: { text: "Success", color: "text-green-400" },
  failed: { text: "Failed", color: "text-red-400" },
  running: { text: "Running", color: "text-yellow-400" },
  pending: { text: "Pending", color: "text-gray-400" },
};

export default function JobDetail({ job, onEdit }: Props) {
  const promptHtml = useMemo(() => {
    if (!job?.prompt) return "";
    return marked.parse(job.prompt, { async: false }) as string;
  }, [job?.prompt]);

  if (!job) {
    return (
      <div className="h-full flex items-center justify-center bg-gray-900 text-gray-500">
        Select a job to view details
      </div>
    );
  }

  const status = statusLabels[job.status];

  return (
    <div className="h-full flex flex-col bg-gray-900">
      <div className="px-6 py-4 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <h1 className="text-lg font-semibold text-gray-100">{job.name}</h1>
          <div className="flex items-center gap-3">
            <span className={`text-sm font-medium ${status.color}`}>
              {status.text}
            </span>
            <button
              onClick={onEdit}
              className="text-sm text-blue-400 hover:text-blue-300 px-2 py-1 rounded border border-gray-600 hover:border-gray-500 transition-colors"
            >
              Edit
            </button>
          </div>
        </div>
        <div className="mt-2 flex gap-6 text-xs text-gray-400">
          <span>
            Schedule:{" "}
            <span className="text-gray-300">
              {formatInterval(job.intervalValue, job.intervalUnit)}
            </span>
          </span>
          <span>
            Starts:{" "}
            <span className="text-gray-300">{formatTime(job.startDate)}</span>
          </span>
          <span>
            <span className={job.active ? "text-green-400" : "text-gray-500"}>
              {job.active ? "Active" : "Inactive"}
            </span>
          </span>
          <span>Next run: {formatTime(job.nextRun)}</span>
          <span>Last run: {formatTime(job.lastRun)}</span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        {job.prompt && (
          <div className="mb-6">
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
              Prompt
            </h3>
            <div
              className="prose prose-invert prose-sm max-w-none"
              dangerouslySetInnerHTML={{ __html: promptHtml }}
            />
          </div>
        )}
        <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
          Run History
        </h3>
        <RunHistory jobId={job.id} />
      </div>
    </div>
  );
}
