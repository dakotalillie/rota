import { describe, expect, it } from "vitest";

import type { Member, TimeSegment } from "./types";
import {
  cn,
  computeOverrideReplacements,
  formatDateTimeRange,
  initials,
} from "./utils";

describe("cn", () => {
  it("joins class strings", () => {
    expect(cn("foo", "bar")).toBe("foo bar");
  });

  it("ignores falsy values", () => {
    expect(cn("foo", undefined, null, false, "bar")).toBe("foo bar");
  });

  it("handles conditional object syntax", () => {
    expect(cn({ foo: true, bar: false })).toBe("foo");
  });

  it("resolves tailwind conflicts — last class wins", () => {
    expect(cn("p-4", "p-2")).toBe("p-2");
  });
});

describe("formatDateTimeRange", () => {
  it("joins start and end with an en-dash", () => {
    const start = new Date("2025-06-01T09:00:00");
    const end = new Date("2025-06-01T17:00:00");
    expect(formatDateTimeRange(start, end)).toBe(
      "Jun 1, 9:00 AM – Jun 1, 5:00 PM",
    );
  });

  it("works across day boundaries", () => {
    const start = new Date("2025-06-01T22:00:00");
    const end = new Date("2025-06-02T06:00:00");
    expect(formatDateTimeRange(start, end)).toBe(
      "Jun 1, 10:00 PM – Jun 2, 6:00 AM",
    );
  });

  it("works across month boundaries", () => {
    const start = new Date("2025-05-31T08:00:00");
    const end = new Date("2025-06-01T08:00:00");
    expect(formatDateTimeRange(start, end)).toBe(
      "May 31, 8:00 AM – Jun 1, 8:00 AM",
    );
  });
});

describe("initials", () => {
  it("returns uppercase initials for a full name", () => {
    expect(initials("Jane Doe")).toBe("JD");
  });

  it("handles a single name", () => {
    expect(initials("Alice")).toBe("A");
  });

  it("handles three names", () => {
    expect(initials("John Paul Jones")).toBe("JPJ");
  });
});

describe("computeOverrideReplacements", () => {
  const alice: Member = {
    id: "mem_1",
    userId: "usr_1",
    name: "Alice Adams",
    email: "alice@example.com",
    color: "bg-violet-500",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950/50",
    textColor: "text-white",
  };

  const bob: Member = {
    id: "mem_2",
    userId: "usr_2",
    name: "Bob Brown",
    email: "bob@example.com",
    color: "bg-sky-500",
    lightColor: "bg-sky-50",
    darkColor: "dark:bg-sky-950/50",
    textColor: "text-white",
  };

  const schedule: TimeSegment[] = [
    {
      start: new Date("2026-04-06T16:00:00Z"),
      end: new Date("2026-04-13T16:00:00Z"),
      member: alice,
      isOverride: false,
    },
    {
      start: new Date("2026-04-13T16:00:00Z"),
      end: new Date("2026-04-20T16:00:00Z"),
      member: bob,
      isOverride: false,
    },
  ];

  it("returns empty array for invalid dates", () => {
    expect(
      computeOverrideReplacements(schedule, "not-a-date", "2026-04-10"),
    ).toEqual([]);
    expect(
      computeOverrideReplacements(schedule, "2026-04-10", "not-a-date"),
    ).toEqual([]);
  });

  it("returns empty array when end is not after start", () => {
    expect(
      computeOverrideReplacements(
        schedule,
        "2026-04-10T09:00",
        "2026-04-09T09:00",
      ),
    ).toEqual([]);
  });

  it("returns empty array for empty strings", () => {
    expect(
      computeOverrideReplacements(schedule, "", "2026-04-10T09:00"),
    ).toEqual([]);
    expect(
      computeOverrideReplacements(schedule, "2026-04-10T09:00", ""),
    ).toEqual([]);
  });

  it("clips overlapping segments to the override window", () => {
    const results = computeOverrideReplacements(
      schedule,
      "2026-04-09T09:00:00Z",
      "2026-04-15T09:00:00Z",
    );

    expect(results).toHaveLength(2);
    expect(results[0].member.id).toBe("mem_1");
    expect(results[0].start).toEqual(new Date("2026-04-09T09:00:00Z"));
    expect(results[0].end).toEqual(new Date("2026-04-13T16:00:00Z"));
    expect(results[1].member.id).toBe("mem_2");
    expect(results[1].start).toEqual(new Date("2026-04-13T16:00:00Z"));
    expect(results[1].end).toEqual(new Date("2026-04-15T09:00:00Z"));
  });

  it("skips non-overlapping segments", () => {
    const results = computeOverrideReplacements(
      schedule,
      "2026-04-14T00:00:00Z",
      "2026-04-16T00:00:00Z",
    );

    expect(results).toHaveLength(1);
    expect(results[0].member.id).toBe("mem_2");
  });
});
