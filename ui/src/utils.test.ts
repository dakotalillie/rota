import { describe, expect, it } from "vitest";

import { cn, initials } from "./utils";

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
