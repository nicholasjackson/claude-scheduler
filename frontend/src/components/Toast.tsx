import { Toast, ToastType } from "../hooks/useToasts";

interface Props {
  toasts: Toast[];
  onDismiss: (id: number) => void;
}

const styles: Record<ToastType, { bg: string; border: string; text: string; dot: string }> = {
  info: {
    bg: "bg-yellow-900/30",
    border: "border-yellow-700",
    text: "text-yellow-300",
    dot: "bg-yellow-400 animate-pulse",
  },
  success: {
    bg: "bg-green-900/30",
    border: "border-green-700",
    text: "text-green-300",
    dot: "bg-green-400",
  },
  error: {
    bg: "bg-red-900/30",
    border: "border-red-700",
    text: "text-red-300",
    dot: "bg-red-400",
  },
};

export default function ToastContainer({ toasts, onDismiss }: Props) {
  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
      {toasts.map((toast) => {
        const s = styles[toast.type];
        return (
          <div
            key={toast.id}
            className={`flex items-center gap-3 px-4 py-3 rounded border ${s.bg} ${s.border} shadow-lg animate-[slideIn_0.2s_ease-out]`}
          >
            <span className={`w-2 h-2 rounded-full shrink-0 ${s.dot}`} />
            <span className={`text-sm ${s.text} flex-1`}>{toast.message}</span>
            <button
              onClick={() => onDismiss(toast.id)}
              className="text-gray-500 hover:text-gray-300 text-xs ml-2"
            >
              &times;
            </button>
          </div>
        );
      })}
    </div>
  );
}
