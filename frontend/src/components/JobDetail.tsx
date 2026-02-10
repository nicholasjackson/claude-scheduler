import { useEffect, useMemo, useState } from "react";
import { marked } from "marked";
import { ScheduledJob, JobRun } from "../types";
import { formatInterval, formatTime } from "../utils";
import { GetRunsForJob, OnEvent, AnswerQuestion } from "../wailsbridge";
import RunHistory from "./RunHistory";

type Tab = "current" | "history";

interface Props {
  job: ScheduledJob | null;
  onEdit: () => void;
  onRunNow: (jobId: string) => void;
}

const statusLabels: Record<string, { text: string; color: string }> = {
  success: { text: "Success", color: "text-green-400" },
  failed: { text: "Failed", color: "text-red-400" },
  running: { text: "Running", color: "text-yellow-400" },
  pending: { text: "Pending", color: "text-gray-400" },
  waiting: { text: "Waiting for Input", color: "text-amber-400" },
};

const statusDot: Record<string, string> = {
  success: "bg-green-400",
  failed: "bg-red-400",
  running: "bg-yellow-400 animate-pulse",
  waiting: "bg-amber-400 animate-pulse",
};

interface QuestionOption {
  label: string;
  description: string;
}

interface QuestionItem {
  question: string;
  header: string;
  options: QuestionOption[];
}

interface QuestionInput {
  questions: QuestionItem[];
}

function parseQuestion(json: string): QuestionInput | null {
  if (!json) return null;
  try {
    const parsed = JSON.parse(json) as QuestionInput;
    if (parsed.questions && parsed.questions.length > 0) return parsed;
  } catch { /* ignore */ }
  return null;
}

function PendingQuestionUI({ jobId, questionJson }: { jobId: string; questionJson: string }) {
  const [answering, setAnswering] = useState(false);
  const question = parseQuestion(questionJson);
  if (!question) return null;

  const handleAnswer = async (answer: string) => {
    setAnswering(true);
    try {
      await AnswerQuestion(jobId, answer);
    } catch (err) {
      console.error("Failed to answer question:", err);
      setAnswering(false);
    }
  };

  return (
    <div className="border-t border-amber-800 bg-amber-950/30 p-4">
      {question.questions.map((q, idx) => (
        <div key={idx} className="mb-4 last:mb-0">
          {q.header && (
            <div className="text-xs font-bold text-amber-400 uppercase tracking-wider mb-1">
              {q.header}
            </div>
          )}
          <p className="text-sm text-gray-200 mb-3">{q.question}</p>
          <div className="flex flex-wrap gap-2">
            {q.options.map((opt, oidx) => (
              <button
                key={oidx}
                disabled={answering}
                onClick={() => handleAnswer(opt.label)}
                className={`px-4 py-2 rounded border text-sm font-medium transition-colors ${
                  answering
                    ? "border-gray-700 text-gray-500 cursor-not-allowed"
                    : "border-amber-600 text-amber-300 hover:bg-amber-900/50 hover:border-amber-500"
                }`}
                title={opt.description}
              >
                {opt.label}
              </button>
            ))}
          </div>
          {answering && (
            <p className="text-xs text-amber-400 mt-2 animate-pulse">Sending answer...</p>
          )}
        </div>
      ))}
    </div>
  );
}

function CurrentRunOutput({ run, jobId }: { run: JobRun; jobId: string }) {
  const html = useMemo(() => {
    if (!run.output) return "";
    return marked.parse(run.output, { async: false }) as string;
  }, [run.output]);

  const dot = statusDot[run.status] ?? "bg-gray-400";
  const statusText = run.status === "running" ? "running…" : run.status === "waiting" ? "waiting for input" : run.status;

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-700">
        <span className={`w-2.5 h-2.5 rounded-full shrink-0 ${dot}`} />
        <span className="text-sm text-gray-300">{formatTime(run.startedAt)}</span>
        <span className="text-xs text-gray-500">{statusText}</span>
      </div>
      {run.output ? (
        <div className="flex-1 overflow-y-auto">
          <div
            className="p-4 text-sm text-gray-300"
            dangerouslySetInnerHTML={{ __html: html }}
          />
          {run.status === "waiting" && run.pendingQuestion && (
            <PendingQuestionUI jobId={jobId} questionJson={run.pendingQuestion} />
          )}
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center">
          <p className="text-sm text-gray-500 italic">
            {run.status === "running" ? "Running…" : "(no output)"}
          </p>
        </div>
      )}
    </div>
  );
}

export default function JobDetail({ job, onEdit, onRunNow }: Props) {
  const [activeTab, setActiveTab] = useState<Tab>("current");
  const [latestRun, setLatestRun] = useState<JobRun | null>(null);

  const promptHtml = useMemo(() => {
    if (!job?.prompt) return "";
    return marked.parse(job.prompt, { async: false }) as string;
  }, [job?.prompt]);

  useEffect(() => {
    if (!job?.id) return;
    const fetchLatest = () => {
      GetRunsForJob(job.id).then((data) => {
        const list = data ?? [];
        setLatestRun(list.length > 0 ? list[0] : null);
      });
    };
    fetchLatest();
    const off = OnEvent("jobs:updated", fetchLatest);
    return off;
  }, [job?.id]);

  if (!job) {
    return (
      <div className="h-full flex items-center justify-center bg-gray-900 text-gray-500">
        Select a job to view details
      </div>
    );
  }

  const status = statusLabels[job.status];

  const tabs: { key: Tab; label: string }[] = [
    { key: "current", label: "Current Run" },
    { key: "history", label: "Run History" },
  ];

  return (
    <div className="h-full flex flex-col bg-gray-900">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <h1 className="text-lg font-semibold text-gray-100">{job.name}</h1>
          <div className="flex items-center gap-3">
            <span className={`text-sm font-medium ${status.color}`}>
              {status.text}
            </span>
            <button
              onClick={() => onRunNow(job.id)}
              disabled={job.status === "running" || job.status === "waiting"}
              className={`text-sm px-2 py-1 rounded border transition-colors ${
                job.status === "running" || job.status === "waiting"
                  ? "text-gray-500 border-gray-700 cursor-not-allowed"
                  : "text-green-400 hover:text-green-300 border-gray-600 hover:border-gray-500"
              }`}
            >
              Run Now
            </button>
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

        {/* Prompt (collapsible) */}
        {job.prompt && (
          <details className="mt-3">
            <summary className="text-xs font-semibold text-gray-500 uppercase tracking-wider cursor-pointer hover:text-gray-400">
              Prompt
            </summary>
            <div
              className="mt-2 prose prose-invert prose-sm max-w-none"
              dangerouslySetInnerHTML={{ __html: promptHtml }}
            />
          </details>
        )}
      </div>

      {/* Tab bar */}
      <div className="flex border-b border-gray-700 px-6">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === tab.key
                ? "text-green-400 border-b-2 border-green-400"
                : "text-gray-400 hover:text-gray-200"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === "current" && (
          latestRun ? (
            <CurrentRunOutput run={latestRun} jobId={job.id} />
          ) : (
            <div className="h-full flex items-center justify-center">
              <p className="text-sm text-gray-500 italic">No runs yet.</p>
            </div>
          )
        )}
        {activeTab === "history" && (
          <div className="h-full overflow-y-auto p-6">
            <RunHistory jobId={job.id} />
          </div>
        )}
      </div>
    </div>
  );
}
