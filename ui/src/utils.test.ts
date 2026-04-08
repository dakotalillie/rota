import { describe, expect, it } from "vitest";

import type { Member } from "./types";
import { buildTimeline, cn, formatDateTimeRange, initials } from "./utils";

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

describe("buildTimeline", () => {
  const members: Member[] = [
    {
      id: "mem_1",
      userId: "usr_1",
      name: "Alice Adams",
      email: "alice@example.com",
      color: "bg-violet-500",
      lightColor: "bg-violet-50",
      darkColor: "dark:bg-violet-950/50",
      textColor: "text-white",
    },
    {
      id: "mem_2",
      userId: "usr_2",
      name: "Bob Brown",
      email: "bob@example.com",
      color: "bg-sky-500",
      lightColor: "bg-sky-50",
      darkColor: "dark:bg-sky-950/50",
      textColor: "text-white",
    },
    {
      id: "mem_3",
      userId: "usr_3",
      name: "Casey Clark",
      email: "casey@example.com",
      color: "bg-emerald-500",
      lightColor: "bg-emerald-50",
      darkColor: "dark:bg-emerald-950/50",
      textColor: "text-white",
    },
  ];

  it("starts the timeline from the scheduled member instead of index zero", () => {
    const timeline = buildTimeline(members, [], 3, "mem_2");

    expect(timeline).toHaveLength(3);
    expect(timeline.map((segment) => segment.member.id)).toEqual([
      "mem_2",
      "mem_3",
      "mem_1",
    ]);
  });
});
