import { ScheduledJob } from "../types";
import JobListItem from "./JobListItem";

interface Props {
  jobs: ScheduledJob[];
  selectedJobId: string | null;
  onSelectJob: (id: string) => void;
  onNewJob: () => void;
}

export default function JobList({ jobs, selectedJobId, onSelectJob, onNewJob }: Props) {
  return (
    <div className="h-full flex flex-col bg-gray-800">
      <div className="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-300 uppercase tracking-wider">
          Scheduled Jobs
        </h2>
        <button
          onClick={onNewJob}
          className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
          title="New Job"
        >
          + New
        </button>
      </div>
      <div className="flex-1 overflow-y-auto">
        {jobs.map((job) => (
          <JobListItem
            key={job.id}
            job={job}
            isSelected={job.id === selectedJobId}
            onSelect={onSelectJob}
          />
        ))}
      </div>
    </div>
  );
}
