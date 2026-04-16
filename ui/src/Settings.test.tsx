// @vitest-environment jsdom

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { useLayoutEffect } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AppStateProvider, useAppState } from "./AppStateContext";
import Settings from "./Settings";
import type { Member, Override } from "./types";

vi.mock("@tanstack/react-router", () => ({
  useParams: () => ({ rotationId: "rot_123" }),
}));

vi.mock("./AddMemberDialog", () => ({
  default: () => null,
}));

vi.mock("./AddOverrideDialog", () => ({
  default: () => null,
}));

vi.mock("./PageHeader", () => ({
  default: ({ title }: { title: string }) => <h1>{title}</h1>,
}));

const staleMembers: Member[] = [
  {
    id: "mem_stale",
    userId: "usr_stale",
    name: "Stale Member",
    email: "stale@example.com",
    color: "bg-violet-500",
    lightColor: "bg-violet-50",
    darkColor: "dark:bg-violet-950/50",
    textColor: "text-white",
  },
];

const staleOverrides: Override[] = [
  {
    id: "ovr_stale",
    start: "2026-04-01T09:00",
    end: "2026-04-02T09:00",
    memberId: "mem_stale",
  },
];

function createDeferredResponse() {
  let resolve!: (value: Response) => void;
  const promise = new Promise<Response>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

function SeedState({
  members,
  overrides,
}: {
  members: Member[];
  overrides: Override[];
}) {
  const { setMembers, setOverrides } = useAppState();

  useLayoutEffect(() => {
    setMembers(members);
    setOverrides(overrides);
  }, [members, overrides, setMembers, setOverrides]);

  return null;
}

function renderSettings(options?: {
  initialMembers?: Member[];
  initialOverrides?: Override[];
}) {
  render(
    <AppStateProvider>
      {options ? (
        <SeedState
          members={options.initialMembers ?? []}
          overrides={options.initialOverrides ?? []}
        />
      ) : null}
      <Settings />
    </AppStateProvider>,
  );
}

function createRotationResponse(withOverrides = true) {
  return {
    data: {
      type: "rotations",
      id: "rot_123",
      attributes: { name: "Platform On-Call" },
      relationships: {
        members: {
          data: [{ id: "mem_1" }, { id: "mem_2" }],
        },
        overrides: {
          data: withOverrides ? [{ id: "ovr_1" }] : [],
        },
      },
    },
    included: [
      {
        type: "members",
        id: "mem_1",
        attributes: { position: 2, color: "sky" },
        relationships: { user: { data: { id: "usr_1" } } },
      },
      {
        type: "users",
        id: "usr_1",
        attributes: { name: "Alice Adams", email: "alice@example.com" },
      },
      {
        type: "members",
        id: "mem_2",
        attributes: { position: 1, color: "violet" },
        relationships: { user: { data: { id: "usr_2" } } },
      },
      {
        type: "users",
        id: "usr_2",
        attributes: { name: "Bob Brown", email: "bob@example.com" },
      },
      ...(withOverrides
        ? [
            {
              type: "overrides",
              id: "ovr_1",
              attributes: {
                start: "2026-04-07T09:00:00Z",
                end: "2026-04-14T09:00:00Z",
              },
              relationships: { member: { data: { id: "mem_2" } } },
            },
          ]
        : []),
    ],
  };
}

const scheduleResponse = { data: [] };

describe("Settings", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("hydrates members and overrides from a single rotation response", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation((url) => {
        const body = String(url).endsWith("/schedule")
          ? scheduleResponse
          : createRotationResponse(true);
        return Promise.resolve(
          new Response(JSON.stringify(body), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          }),
        );
      });

    renderSettings();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/rotations/rot_123");
    expect(fetchMock).toHaveBeenCalledWith("/api/rotations/rot_123/schedule");

    await screen.findByText("Apr 7, 9:00 AM – Apr 14, 9:00 AM");
    expect(screen.getAllByText("Alice Adams").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Bob Brown").length).toBeGreaterThan(1);
    expect(screen.queryByText("No overrides scheduled.")).toBeNull();
  });

  it("shows an empty overrides state when the response has no overrides", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation((url) => {
        const body = String(url).endsWith("/schedule")
          ? scheduleResponse
          : createRotationResponse(false);
        return Promise.resolve(
          new Response(JSON.stringify(body), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          }),
        );
      });

    renderSettings();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    await screen.findByText("Alice Adams");
    expect(screen.getByText("No overrides scheduled.")).toBeTruthy();
    expect(screen.queryByText("Apr 7, 9:00 AM – Apr 14, 9:00 AM")).toBeNull();
  });

  it("clears stale members and overrides when hydration fails", async () => {
    const deferred = createDeferredResponse();
    vi.spyOn(globalThis, "fetch").mockReturnValue(deferred.promise);

    renderSettings({
      initialMembers: staleMembers,
      initialOverrides: staleOverrides,
    });

    await screen.findAllByText("Stale Member");
    expect(screen.getByText("Apr 1, 9:00 AM – Apr 2, 9:00 AM")).toBeTruthy();

    deferred.resolve(
      new Response(
        JSON.stringify({ errors: [{ detail: "Rotation not found" }] }),
        {
          status: 404,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );

    await waitFor(() => {
      expect(screen.getByText("No members yet.")).toBeTruthy();
      expect(screen.getByText("No overrides scheduled.")).toBeTruthy();
    });

    expect(screen.queryByText("Stale Member")).toBeNull();
    expect(screen.queryByText("Apr 1, 9:00 AM – Apr 2, 9:00 AM")).toBeNull();
  });
});
