import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

import { colorsForName } from "./colorPalette";
import type { Member, TimeSegment } from "./types";

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

export interface ApiMember {
  type: "members";
  id: string;
  attributes: { position: number; color: string };
  relationships: { user: { data: { type: "users"; id: string } } };
}

export interface ApiUser {
  type: "users";
  id: string;
  attributes: { name: string; email: string };
}

export interface ApiScheduleBlock {
  type: "scheduleBlocks";
  id: string;
  attributes: { start: string; end: string; isOverride: boolean };
  relationships: { member: { data: { type: "members"; id: string } } };
}

export function buildTimelineFromSchedule(
  data: ApiScheduleBlock[],
  included: (ApiMember | ApiUser)[] | undefined,
): TimeSegment[] {
  const userMap = new Map<string, ApiUser>();
  const memberMap = new Map<string, ApiMember>();

  for (const item of included ?? []) {
    if (item.type === "users") userMap.set(item.id, item);
    if (item.type === "members") memberMap.set(item.id, item);
  }

  const members = new Map<string, Member>();

  for (const block of data) {
    const memberId = block.relationships.member.data.id;
    if (!members.has(memberId)) {
      const apiMember = memberMap.get(memberId);
      const userId = apiMember?.relationships.user.data.id;
      if (!userId) continue;
      const user = userMap.get(userId);
      if (!user) continue;
      members.set(memberId, {
        id: memberId,
        userId,
        name: user.attributes.name,
        email: user.attributes.email,
        ...colorsForName(apiMember.attributes.color),
      });
    }
  }

  return data.map((block) => {
    const memberId = block.relationships.member.data.id;
    const member = members.get(memberId) ?? {
      id: memberId,
      userId: "",
      name: "Unknown",
      email: "",
      ...colorsForName(""),
    };
    return {
      start: new Date(block.attributes.start),
      end: new Date(block.attributes.end),
      member,
      isOverride: block.attributes.isOverride,
    };
  });
}

/**
 * Given a prospective override window [previewStart, previewEnd), compute which
 * segments from the backend-computed schedule would be displaced. Returns those
 * segments clipped to the override window.
 */
export function computeOverrideReplacements(
  schedule: TimeSegment[],
  previewStart: string,
  previewEnd: string,
): TimeSegment[] {
  if (!previewStart || !previewEnd) return [];
  const start = new Date(previewStart);
  const end = new Date(previewEnd);
  if (isNaN(start.getTime()) || isNaN(end.getTime()) || end <= start) return [];

  return schedule
    .filter((seg) => seg.start < end && seg.end > start)
    .map((seg) => ({
      ...seg,
      start: seg.start < start ? new Date(start) : seg.start,
      end: seg.end > end ? new Date(end) : seg.end,
    }));
}
