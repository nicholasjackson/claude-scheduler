import { IntervalUnit } from "./types";

const unitLabels: Record<IntervalUnit, { singular: string; plural: string }> = {
  minutes: { singular: "minute", plural: "minutes" },
  hours: { singular: "hour", plural: "hours" },
  days: { singular: "day", plural: "days" },
  weeks: { singular: "week", plural: "weeks" },
};

export function formatInterval(value: number, unit: IntervalUnit): string {
  const label = unitLabels[unit];
  return `Every ${value} ${value === 1 ? label.singular : label.plural}`;
}

export function formatTime(value: string): string {
  if (!value) return "—";
  const d = new Date(value);
  if (isNaN(d.getTime())) return "—";
  return d.toLocaleString();
}
