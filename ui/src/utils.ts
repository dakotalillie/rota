import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

import type { Member, Override, TimeSegment } from "./types";

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

const dateTimeFormat = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
  hour: "numeric",
  minute: "2-digit",
});

export function formatDateTimeRange(start: Date, end: Date): string {
  return `${dateTimeFormat.format(start)} – ${dateTimeFormat.format(end)}`;
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
 * 3. For each sub-interval, checking whether an override covers it; if not, using the rotation member
 * 4. Merging adjacent segments with the same member
 */
export function buildTimeline(
  members: Member[],
  overrides: Override[],
  weeksCount: number,
  scheduledMemberId?: string | null,
): TimeSegment[] {
  if (members.length === 0) return [];

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
  const startIdx = scheduledMemberId
    ? Math.max(
        members.findIndex((member) => member.id === scheduledMemberId),
        0,
      )
    : 0;

  // Build raw segments
  const raw: TimeSegment[] = [];
  for (let i = 0; i < boundaries.length - 1; i++) {
    const segStart = new Date(boundaries[i]);
    const segEnd = new Date(boundaries[i + 1]);
    const midMs = (boundaries[i] + boundaries[i + 1]) / 2;

    // Find the most-recently-added override that covers this segment (last one wins)
    let overrideMember: Member | undefined;
    for (let j = overrides.length - 1; j >= 0; j--) {
      const ov = overrides[j];
      const ovStart = new Date(ov.start).getTime();
      const ovEnd = new Date(ov.end).getTime();
      if (ovStart <= midMs && ovEnd > midMs) {
        overrideMember = members.find((m) => m.id === ov.memberId);
        break;
      }
    }

    if (overrideMember) {
      raw.push({
        start: segStart,
        end: segEnd,
        member: overrideMember,
        isOverride: true,
      });
    } else {
      // Which week index does this fall in?
      const weekIndex = Math.floor(
        (boundaries[i] - start.getTime()) / (7 * 24 * 60 * 60 * 1000),
      );
      const member = members[(startIdx + weekIndex) % members.length];
      raw.push({ start: segStart, end: segEnd, member, isOverride: false });
    }
  }

  // Merge adjacent segments with the same member and override status
  const merged: TimeSegment[] = [];
  for (const seg of raw) {
    const last = merged[merged.length - 1];
    if (
      last &&
      last.member.id === seg.member.id &&
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
  members: Member[],
  baseOverrides: Override[],
  previewStart: string,
  previewEnd: string,
  scheduledMemberId?: string | null,
): TimeSegment[] {
  if (members.length === 0 || !previewStart || !previewEnd) return [];
  const start = new Date(previewStart);
  const end = new Date(previewEnd);
  if (isNaN(start.getTime()) || isNaN(end.getTime()) || end <= start) return [];

  const timeline = buildTimeline(members, baseOverrides, 8, scheduledMemberId);
  return timeline
    .filter((seg) => seg.start < end && seg.end > start)
    .map((seg) => ({
      ...seg,
      start: seg.start < start ? new Date(start) : seg.start,
      end: seg.end > end ? new Date(end) : seg.end,
    }));
}
