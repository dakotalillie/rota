import { describe, expect, it } from "vitest";

import { initials } from "./utils";

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
