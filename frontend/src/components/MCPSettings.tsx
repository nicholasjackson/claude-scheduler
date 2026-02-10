import { useEffect, useState } from "react";
import { MCPServer, MCPServerType } from "../types";
import {
  GetMCPServers,
  CreateMCPServer,
  UpdateMCPServer,
  DeleteMCPServer,
} from "../wailsbridge";

const emptyServer: MCPServer = {
  id: "",
  name: "",
  type: "http",
  url: "",
  command: "",
  args: "[]",
  env: "{}",
  headers: "{}",
};

interface Props {
  onClose: () => void;
}

export default function MCPSettings({ onClose }: Props) {
  const [servers, setServers] = useState<MCPServer[]>([]);
  const [editing, setEditing] = useState<MCPServer | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = () => {
    GetMCPServers().then((data) => setServers(data ?? []));
  };

  useEffect(() => {
    refresh();
  }, []);

  const handleSave = async (srv: MCPServer) => {
    setError(null);
    try {
      if (srv.id) {
        await UpdateMCPServer(srv);
      } else {
        await CreateMCPServer(srv);
      }
      setEditing(null);
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const handleDelete = async (id: string) => {
    setError(null);
    try {
      await DeleteMCPServer(id);
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  return (
    <div className="h-full flex flex-col bg-gray-900">
      <div className="px-6 py-4 border-b border-gray-700 flex items-center justify-between">
        <h1 className="text-lg font-semibold text-gray-100">MCP Servers</h1>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-200 transition-colors text-sm"
        >
          Back
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        {error && (
          <p className="mb-4 text-sm text-red-400 bg-red-900/20 border border-red-800 rounded px-3 py-2">
            {error}
          </p>
        )}

        {editing ? (
          <ServerForm
            server={editing}
            onSave={handleSave}
            onCancel={() => {
              setEditing(null);
              setError(null);
            }}
          />
        ) : (
          <>
            <button
              onClick={() => setEditing({ ...emptyServer })}
              className="mb-4 px-4 py-2 rounded text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 transition-colors"
            >
              Add Server
            </button>

            {servers.length === 0 ? (
              <p className="text-sm text-gray-500 italic">
                No MCP servers configured.
              </p>
            ) : (
              <div className="space-y-2">
                {servers.map((srv) => (
                  <div
                    key={srv.id}
                    className="border border-gray-700 rounded px-4 py-3 flex items-center justify-between"
                  >
                    <div>
                      <p className="text-sm font-medium text-gray-100">
                        {srv.name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {srv.type === "http" ? srv.url : `${srv.command}`}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <span className="text-xs px-2 py-0.5 rounded bg-gray-800 text-gray-400 border border-gray-700">
                        {srv.type}
                      </span>
                      <button
                        onClick={() => setEditing({ ...srv })}
                        className="text-xs text-blue-400 hover:text-blue-300"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(srv.id)}
                        className="text-xs text-red-400 hover:text-red-300"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function KeyValueEditor({
  value,
  onChange,
}: {
  value: string;
  onChange: (json: string) => void;
}) {
  const parse = (v: string): { key: string; value: string }[] => {
    try {
      const obj = JSON.parse(v);
      if (obj && typeof obj === "object" && !Array.isArray(obj)) {
        return Object.entries(obj).map(([k, val]) => ({ key: k, value: String(val) }));
      }
    } catch { /* ignore */ }
    return [];
  };

  const [entries, setEntries] = useState(parse(value));

  const sync = (updated: { key: string; value: string }[]) => {
    setEntries(updated);
    const obj: Record<string, string> = {};
    for (const e of updated) {
      if (e.key.trim()) obj[e.key.trim()] = e.value;
    }
    onChange(JSON.stringify(obj));
  };

  const update = (idx: number, field: "key" | "value", val: string) => {
    const next = entries.map((e, i) => (i === idx ? { ...e, [field]: val } : e));
    sync(next);
  };

  const remove = (idx: number) => {
    sync(entries.filter((_, i) => i !== idx));
  };

  const add = () => {
    sync([...entries, { key: "", value: "" }]);
  };

  const rowInputClass =
    "bg-gray-800 border border-gray-600 rounded px-2 py-1.5 text-sm text-gray-100 focus:border-blue-500 focus:outline-none";

  return (
    <div className="space-y-2">
      {entries.map((entry, idx) => (
        <div key={idx} className="flex gap-2 items-center">
          <input
            type="text"
            value={entry.key}
            onChange={(e) => update(idx, "key", e.target.value)}
            placeholder="Key"
            className={rowInputClass + " flex-1"}
          />
          <input
            type="text"
            value={entry.value}
            onChange={(e) => update(idx, "value", e.target.value)}
            placeholder="Value"
            className={rowInputClass + " flex-[2]"}
          />
          <button
            onClick={() => remove(idx)}
            className="text-red-400 hover:text-red-300 text-sm px-1.5 py-1 shrink-0"
          >
            x
          </button>
        </div>
      ))}
      <button
        onClick={add}
        className="text-xs text-blue-400 hover:text-blue-300"
      >
        + Add
      </button>
    </div>
  );
}

function ListEditor({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (json: string) => void;
  placeholder?: string;
}) {
  const parse = (v: string): string[] => {
    try {
      const arr = JSON.parse(v);
      if (Array.isArray(arr)) return arr.map(String);
    } catch { /* ignore */ }
    return [];
  };

  const [items, setItems] = useState(parse(value));

  const sync = (updated: string[]) => {
    setItems(updated);
    onChange(JSON.stringify(updated));
  };

  const update = (idx: number, val: string) => {
    sync(items.map((item, i) => (i === idx ? val : item)));
  };

  const remove = (idx: number) => {
    sync(items.filter((_, i) => i !== idx));
  };

  const add = () => {
    sync([...items, ""]);
  };

  const rowInputClass =
    "bg-gray-800 border border-gray-600 rounded px-2 py-1.5 text-sm text-gray-100 focus:border-blue-500 focus:outline-none";

  return (
    <div className="space-y-2">
      {items.map((item, idx) => (
        <div key={idx} className="flex gap-2 items-center">
          <input
            type="text"
            value={item}
            onChange={(e) => update(idx, e.target.value)}
            placeholder={placeholder ?? "Value"}
            className={rowInputClass + " flex-1"}
          />
          <button
            onClick={() => remove(idx)}
            className="text-red-400 hover:text-red-300 text-sm px-1.5 py-1 shrink-0"
          >
            x
          </button>
        </div>
      ))}
      <button
        onClick={add}
        className="text-xs text-blue-400 hover:text-blue-300"
      >
        + Add
      </button>
    </div>
  );
}

function ServerForm({
  server,
  onSave,
  onCancel,
}: {
  server: MCPServer;
  onSave: (srv: MCPServer) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(server.name);
  const [type, setType] = useState<MCPServerType>(server.type);
  const [url, setUrl] = useState(server.url);
  const [command, setCommand] = useState(server.command);
  const [args, setArgs] = useState(server.args);
  const [env, setEnv] = useState(server.env);
  const [headers, setHeaders] = useState(server.headers);

  const handleSubmit = () => {
    onSave({
      id: server.id,
      name: name.trim(),
      type,
      url: type === "http" ? url.trim() : "",
      command: type === "stdio" ? command.trim() : "",
      args: type === "stdio" ? args : "[]",
      env,
      headers: type === "http" ? headers : "{}",
    });
  };

  const inputClass =
    "w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm text-gray-100 focus:border-blue-500 focus:outline-none";
  const labelClass =
    "block text-xs font-semibold text-gray-400 uppercase tracking-wider mb-1.5";

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-semibold text-gray-200">
        {server.id ? "Edit Server" : "New Server"}
      </h2>

      <div>
        <label className={labelClass}>Name</label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="my-server"
          className={inputClass}
        />
      </div>

      <div>
        <label className={labelClass}>Type</label>
        <select
          value={type}
          onChange={(e) => setType(e.target.value as MCPServerType)}
          className={inputClass}
          style={{ colorScheme: "dark" }}
        >
          <option value="http">HTTP</option>
          <option value="stdio">Stdio</option>
        </select>
      </div>

      {type === "http" ? (
        <>
          <div>
            <label className={labelClass}>URL</label>
            <input
              type="text"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://api.example.com/mcp"
              className={inputClass}
            />
          </div>
          <div>
            <label className={labelClass}>Headers</label>
            <KeyValueEditor value={headers} onChange={setHeaders} />
          </div>
        </>
      ) : (
        <>
          <div>
            <label className={labelClass}>Command</label>
            <input
              type="text"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              placeholder="npx"
              className={inputClass}
            />
          </div>
          <div>
            <label className={labelClass}>Args</label>
            <ListEditor value={args} onChange={setArgs} placeholder="-y" />
          </div>
        </>
      )}

      <div>
        <label className={labelClass}>Environment Variables</label>
        <KeyValueEditor value={env} onChange={setEnv} />
      </div>

      <div className="flex gap-3 pt-2">
        <button
          onClick={onCancel}
          className="px-4 py-2 rounded text-sm font-medium bg-gray-700 text-gray-300 hover:bg-gray-600 transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={handleSubmit}
          className="px-4 py-2 rounded text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 transition-colors"
        >
          Save
        </button>
      </div>
    </div>
  );
}
