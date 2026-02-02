import { useEffect, useState } from "react";
import { ScheduledJob, IntervalUnit, MCPServer } from "../types";
import { GetMCPServers, GetMCPServersForJob, SetJobMCPServers } from "../wailsbridge";

interface Props {
  job: ScheduledJob | null;
  onSave: (job: ScheduledJob, mcpServerIds: string[]) => void;
  onCancel: () => void;
  saveError?: string | null;
}

export default function JobForm({ job, onSave, onCancel, saveError }: Props) {
  const [name, setName] = useState(job?.name ?? "");
  const [startDate, setStartDate] = useState(job?.startDate ?? "");
  const [intervalValue, setIntervalValue] = useState(job?.intervalValue ?? 1);
  const [intervalUnit, setIntervalUnit] = useState<IntervalUnit>(
    job?.intervalUnit ?? "hours"
  );
  const [prompt, setPrompt] = useState(job?.prompt ?? "");
  const [active, setActive] = useState(job?.active ?? true);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // MCP server selection state.
  const [allServers, setAllServers] = useState<MCPServer[]>([]);
  const [selectedServerIds, setSelectedServerIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    GetMCPServers().then((data) => setAllServers(data ?? []));
    if (job?.id) {
      GetMCPServersForJob(job.id).then((data) => {
        const ids = (data ?? []).map((s) => s.id);
        setSelectedServerIds(new Set(ids));
      });
    }
  }, [job?.id]);

  const toggleServer = (id: string) => {
    setSelectedServerIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleSave = () => {
    const errs: Record<string, string> = {};
    if (!name.trim()) errs.name = "Name is required";
    if (!startDate) errs.startDate = "Start date is required";
    if (intervalValue <= 0) errs.intervalValue = "Interval must be greater than 0";
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }

    onSave(
      {
        id: job?.id ?? "",
        name: name.trim(),
        startDate,
        intervalValue,
        intervalUnit,
        prompt,
        active,
        nextRun: job?.nextRun ?? "",
        lastRun: job?.lastRun ?? "",
        status: job?.status ?? "pending",
        output: job?.output ?? "",
      },
      Array.from(selectedServerIds)
    );
  };

  return (
    <div className="h-full flex flex-col bg-gray-900">
      <div className="px-6 py-4 border-b border-gray-700">
        <h1 className="text-lg font-semibold text-gray-100">
          {job ? "Edit Job" : "New Job"}
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto p-6 space-y-5">
        <div>
          <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5">
            Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              setErrors((prev) => ({ ...prev, name: "" }));
            }}
            placeholder="Job name"
            className="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm text-gray-100 focus:border-blue-500 focus:outline-none"
          />
          {errors.name && (
            <p className="mt-1 text-xs text-red-400">{errors.name}</p>
          )}
        </div>

        <div>
          <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5">
            Start Date
          </label>
          <input
            type="datetime-local"
            value={startDate}
            onChange={(e) => {
              setStartDate(e.target.value);
              setErrors((prev) => ({ ...prev, startDate: "" }));
            }}
            className="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm text-gray-100 focus:border-blue-500 focus:outline-none"
          />
          {errors.startDate && (
            <p className="mt-1 text-xs text-red-400">{errors.startDate}</p>
          )}
        </div>

        <div>
          <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5">
            Repeat Every
          </label>
          <div className="flex gap-2">
            <input
              type="number"
              min={1}
              value={intervalValue}
              onChange={(e) => {
                setIntervalValue(parseInt(e.target.value, 10) || 1);
                setErrors((prev) => ({ ...prev, intervalValue: "" }));
              }}
              className="w-24 bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm text-gray-100 focus:border-blue-500 focus:outline-none"
            />
            <select
              value={intervalUnit}
              onChange={(e) => setIntervalUnit(e.target.value as IntervalUnit)}
              className="flex-1 bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm text-gray-100 focus:border-blue-500 focus:outline-none"
            >
              <option value="minutes">Minutes</option>
              <option value="hours">Hours</option>
              <option value="days">Days</option>
              <option value="weeks">Weeks</option>
            </select>
          </div>
          {errors.intervalValue && (
            <p className="mt-1 text-xs text-red-400">{errors.intervalValue}</p>
          )}
        </div>

        <div>
          <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5">
            Prompt
          </label>
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={6}
            placeholder="Enter the Claude instruction for this job..."
            className="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm text-gray-100 focus:border-blue-500 focus:outline-none resize-y"
          />
        </div>

        {allServers.length > 0 && (
          <div>
            <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5">
              MCP Servers
            </label>
            <div className="space-y-1.5">
              {allServers.map((srv) => (
                <label
                  key={srv.id}
                  className="flex items-center gap-2 px-3 py-2 rounded border border-gray-700 hover:border-gray-600 cursor-pointer transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={selectedServerIds.has(srv.id)}
                    onChange={() => toggleServer(srv.id)}
                    className="rounded border-gray-600 bg-gray-800 text-blue-500 focus:ring-blue-500 focus:ring-offset-0"
                  />
                  <span className="text-sm text-gray-200">{srv.name}</span>
                  <span className="text-xs px-1.5 py-0.5 rounded bg-gray-800 text-gray-500 border border-gray-700">
                    {srv.type}
                  </span>
                </label>
              ))}
            </div>
            <p className="mt-1.5 text-xs text-gray-600">
              Selected servers' tools will be available to this job.
            </p>
          </div>
        )}

        <div>
          <label className="block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5">
            Status
          </label>
          <button
            type="button"
            onClick={() => setActive(!active)}
            className={`inline-flex items-center gap-2 px-3 py-1.5 rounded text-sm font-medium transition-colors ${
              active
                ? "bg-green-900/40 text-green-400 border border-green-700"
                : "bg-gray-800 text-gray-400 border border-gray-600"
            }`}
          >
            <span
              className={`w-2 h-2 rounded-full ${
                active ? "bg-green-400" : "bg-gray-500"
              }`}
            />
            {active ? "Active" : "Inactive"}
          </button>
        </div>
      </div>

      <div className="px-6 py-4 border-t border-gray-700 space-y-3">
        {saveError && (
          <p className="text-sm text-red-400 bg-red-900/20 border border-red-800 rounded px-3 py-2">
            Failed to save: {saveError}
          </p>
        )}
      </div>
      <div className="px-6 pb-4 flex gap-3 justify-end">
        <button
          onClick={onCancel}
          className="px-4 py-2 rounded text-sm font-medium bg-gray-700 text-gray-300 hover:bg-gray-600 transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={handleSave}
          className="px-4 py-2 rounded text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 transition-colors"
        >
          Save
        </button>
      </div>
    </div>
  );
}
