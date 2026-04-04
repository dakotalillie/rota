import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import type { Engineer, Override, TimeSegment } from "./types";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function initials(name: string) {
  return name
    .split(" ")
    .map((p) => p[0])
    .join("")
    .toUpperCase();
}

const fmtDateTime = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
  hour: "numeric",
  minute: "2-digit",
});

function isMidnight(date: Date) {
  return (
    date.getHours() === 0 && date.getMinutes() === 0 && date.getSeconds() === 0
  );
}

/**
 * Format a half-open segment [start, end).
 * If both boundaries are at midnight, show as "Apr 1 – Apr 7" (end is exclusive, so display end-1 day).
 * If either has a time component, show full datetimes.
 */
export function formatSegmentRange(start: Date, end: Date): string {
  if (isMidnight(start) && isMidnight(end)) {
    const displayEnd = new Date(end);
    displayEnd.setDate(displayEnd.getDate() - 1);
    return `${fmtDateTime.format(start)} – ${fmtDateTime.format(displayEnd)}`;
  }
  return `${fmtDateTime.format(start)} – ${fmtDateTime.format(end)}`;
}

/** Format an override's stored datetime-local strings for display. */
export function formatOverrideRange(start: string, end: string): string {
  return `${fmtDateTime.format(new Date(start))} – ${fmtDateTime.format(new Date(end))}`;
}

function mondayOf(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  d.setDate(d.getDate() + diff);
  d.setHours(0, 0, 0, 0);
  return d;
}

/**
 * Build a timeline of non-overlapping segments by:
 * 1. Generating base week segments from the rotation
 * 2. Collecting all time boundaries (week starts + override starts/ends)
 * 3. For each sub-interval, checking whether an override covers it; if not, using the rotation engineer
 * 4. Merging adjacent segments with the same engineer
 */
export function buildTimeline(
  engineers: Engineer[],
  overrides: Override[],
  weeksCount: number,
): TimeSegment[] {
  if (engineers.length === 0) return [];

  const start = mondayOf(new Date());
  const endTime = new Date(start);
  endTime.setDate(endTime.getDate() + weeksCount * 7);

  // Collect all relevant time boundaries within the window
  const boundarySet = new Set<number>();
  boundarySet.add(start.getTime());
  boundarySet.add(endTime.getTime());

  // Add a boundary for each week
  const d = new Date(start);
  for (let i = 1; i < weeksCount; i++) {
    d.setDate(d.getDate() + 7);
    boundarySet.add(d.getTime());
  }

  // Add override boundaries, clamped to the window
  for (const ov of overrides) {
    const ovStart = new Date(ov.start).getTime();
    const ovEnd = new Date(ov.end).getTime();
    if (ovEnd <= start.getTime() || ovStart >= endTime.getTime()) continue;
    boundarySet.add(Math.max(ovStart, start.getTime()));
    boundarySet.add(Math.min(ovEnd, endTime.getTime()));
  }

  const boundaries = [...boundarySet].sort((a, b) => a - b);

  // Build raw segments
  const raw: TimeSegment[] = [];
  for (let i = 0; i < boundaries.length - 1; i++) {
    const segStart = new Date(boundaries[i]);
    const segEnd = new Date(boundaries[i + 1]);
    const midMs = (boundaries[i] + boundaries[i + 1]) / 2;

    // Find the most-recently-added override that covers this segment (last one wins)
    let overrideEngineer: Engineer | undefined;
    for (let j = overrides.length - 1; j >= 0; j--) {
      const ov = overrides[j];
      const ovStart = new Date(ov.start).getTime();
      const ovEnd = new Date(ov.end).getTime();
      if (ovStart <= midMs && ovEnd > midMs) {
        overrideEngineer = engineers.find((e) => e.id === ov.engineerId);
        break;
      }
    }

    if (overrideEngineer) {
      raw.push({
        start: segStart,
        end: segEnd,
        engineer: overrideEngineer,
        isOverride: true,
      });
    } else {
      // Which week index does this fall in?
      const weekIndex = Math.floor(
        (boundaries[i] - start.getTime()) / (7 * 24 * 60 * 60 * 1000),
      );
      const engineer = engineers[weekIndex % engineers.length];
      raw.push({ start: segStart, end: segEnd, engineer, isOverride: false });
    }
  }

  // Merge adjacent segments with the same engineer and override status
  const merged: TimeSegment[] = [];
  for (const seg of raw) {
    const last = merged[merged.length - 1];
    if (
      last &&
      last.engineer.id === seg.engineer.id &&
      last.isOverride === seg.isOverride
    ) {
      last.end = seg.end;
    } else {
      merged.push({ ...seg });
    }
  }

  return merged;
}

/**
 * Given a prospective override window [previewStart, previewEnd), compute which
 * segments from the baseline schedule (built without that override) would be
 * displaced. Returns those segments clipped to the override window.
 */
export function computeOverrideReplacements(
  engineers: Engineer[],
  baseOverrides: Override[],
  previewStart: string,
  previewEnd: string,
): TimeSegment[] {
  if (engineers.length === 0 || !previewStart || !previewEnd) return [];
  const start = new Date(previewStart);
  const end = new Date(previewEnd);
  if (isNaN(start.getTime()) || isNaN(end.getTime()) || end <= start) return [];

  const timeline = buildTimeline(engineers, baseOverrides, 8);
  return timeline
    .filter((seg) => seg.start < end && seg.end > start)
    .map((seg) => ({
      ...seg,
      start: seg.start < start ? new Date(start) : seg.start,
      end: seg.end > end ? new Date(end) : seg.end,
    }));
}

export const inputClass =
  "w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 transition-shadow";
