import { ScheduledJob, JobStatus } from "../types";
import { formatInterval, formatTime } from "../utils";

interface Props {
  job: ScheduledJob;
  isSelected: boolean;
  onSelect: (id: string) => void;
}

const statusColors: Record<JobStatus, string> = {
  success: "bg-green-500",
  failed: "bg-red-500",
  running: "bg-yellow-500 animate-pulse",
  pending: "bg-gray-500",
  waiting: "bg-amber-500 animate-pulse",
};

export default function JobListItem({ job, isSelected, onSelect }: Props) {
  return (
    <button
      onClick={() => onSelect(job.id)}
      className={`w-full text-left px-4 py-3 border-b border-gray-700 hover:bg-gray-700 transition-colors ${
        isSelected ? "bg-gray-700" : ""
      } ${!job.active ? "opacity-50" : ""}`}
    >
      <div className="flex items-center justify-between">
        <span className="font-medium text-sm text-gray-100 truncate">
          {job.name}
        </span>
        <span
          className={`w-2.5 h-2.5 rounded-full shrink-0 ml-2 ${statusColors[job.status]}`}
          title={job.status}
        />
      </div>
      <div className="mt-1 text-xs text-gray-400">
        <span>{formatInterval(job.intervalValue, job.intervalUnit)}</span>
        <span className="mx-1">|</span>
        <span>Last: {formatTime(job.lastRun)}</span>
      </div>
    </button>
  );
}
