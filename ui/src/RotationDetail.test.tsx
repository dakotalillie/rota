// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import RotationDetail from "./RotationDetail";

vi.mock("@tanstack/react-router", () => ({
  Link: ({ children }: { children?: React.ReactNode }) => children ?? null,
  useParams: () => ({ rotationId: "rot_123" }),
}));

type ScheduleBlock = {
  id: string;
  attributes: {
    start: string;
    end: string;
    isOverride: boolean;
  };
  relationships: {
    member: {
      data: { type: "members"; id: string };
    };
  };
};

function mockRotationRequests(scheduleBlocks: ScheduleBlock[]) {
  return vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
    const url = typeof input === "string" ? input : (input as Request).url;

    if (url === "/api/rotations/rot_123") {
      return Promise.resolve(
        new Response(
          JSON.stringify({
            data: { attributes: { name: "Primary Rotation" } },
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      );
    }

    if (url === "/api/rotations/rot_123/schedule") {
      return Promise.resolve(
        new Response(
          JSON.stringify({
            data: scheduleBlocks.map((block) => ({
              type: "scheduleBlocks",
              ...block,
            })),
            included: [
              {
                type: "members",
                id: "mem_1",
                attributes: { position: 1, color: "violet" },
                relationships: {
                  user: { data: { type: "users", id: "usr_1" } },
                },
              },
              {
                type: "members",
                id: "mem_2",
                attributes: { position: 2, color: "sky" },
                relationships: {
                  user: { data: { type: "users", id: "usr_2" } },
                },
              },
              {
                type: "users",
                id: "usr_1",
                attributes: {
                  name: "Alice Adams",
                  email: "alice@example.com",
                },
              },
              {
                type: "users",
                id: "usr_2",
                attributes: {
                  name: "Bob Brown",
                  email: "bob@example.com",
                },
              },
            ],
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      );
    }

    return Promise.reject(new Error(`Unhandled request: ${url}`));
  });
}

describe("RotationDetail", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("shows override badges in both the hero and schedule when the active block is an override", async () => {
    const now = new Date();
    const currentBlockStart = new Date(now.getTime() - 60 * 60 * 1000);
    const currentBlockEnd = new Date(now.getTime() + 60 * 60 * 1000);
    const nextBlockEnd = new Date(now.getTime() + 2 * 60 * 60 * 1000);

    mockRotationRequests([
      {
        id: "blk_1",
        attributes: {
          start: currentBlockStart.toISOString(),
          end: currentBlockEnd.toISOString(),
          isOverride: true,
        },
        relationships: { member: { data: { type: "members", id: "mem_1" } } },
      },
      {
        id: "blk_2",
        attributes: {
          start: currentBlockEnd.toISOString(),
          end: nextBlockEnd.toISOString(),
          isOverride: false,
        },
        relationships: { member: { data: { type: "members", id: "mem_2" } } },
      },
    ]);

    render(<RotationDetail />);

    await screen.findByText("Primary Rotation");

    expect(screen.getAllByText("Override")).toHaveLength(2);
  });

  it("omits override badges when no schedule blocks are overrides", async () => {
    mockRotationRequests([
      {
        id: "blk_1",
        attributes: {
          start: "2026-04-07T00:00:00Z",
          end: "2026-04-08T00:00:00Z",
          isOverride: false,
        },
        relationships: { member: { data: { type: "members", id: "mem_1" } } },
      },
    ]);

    render(<RotationDetail />);

    await screen.findByText("Primary Rotation");

    expect(screen.queryByText("Override")).toBeNull();
  });
});
