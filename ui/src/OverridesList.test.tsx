// @vitest-environment jsdom

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLayoutEffect } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AppStateProvider, useAppState } from "./AppStateContext";
import OverridesList from "./OverridesList";
import type { Member, Override } from "./types";

const useParamsMock = vi.fn(() => ({ rotationId: "rot_123" }));

vi.mock("@tanstack/react-router", () => ({
  useParams: () => useParamsMock(),
}));

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
];

const overrides: Override[] = [
  {
    id: "ovr_1",
    start: "2026-04-07T09:00",
    end: "2026-04-08T09:00",
    memberId: "mem_1",
  },
  {
    id: "ovr_2",
    start: "2026-04-08T09:00",
    end: "2026-04-09T09:00",
    memberId: "mem_2",
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
  scheduledMemberId,
}: {
  scheduledMemberId: string | null;
}) {
  const { setScheduledMemberId } = useAppState();

  useLayoutEffect(() => {
    setScheduledMemberId(scheduledMemberId);
  }, [scheduledMemberId, setScheduledMemberId]);

  return null;
}

describe("OverridesList", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    useParamsMock.mockReturnValue({ rotationId: "rot_123" });
  });

  function renderComponent(setOverrides = vi.fn()) {
    render(
      <AppStateProvider>
        <SeedState scheduledMemberId="mem_2" />
        <OverridesList
          members={members}
          overrides={overrides}
          setOverrides={setOverrides}
        />
      </AppStateProvider>,
    );

    return { setOverrides };
  }

  it("deletes an override after a successful response", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response(null, { status: 204 }));
    const { setOverrides } = renderComponent();

    await user.click(
      screen.getAllByRole("button", { name: "Remove override" })[0],
    );

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/rotations/rot_123/overrides/ovr_1",
      { method: "DELETE" },
    );
    expect(setOverrides).toHaveBeenCalledWith([overrides[1]]);
  });

  it("renders the server error detail and preserves state when deletion fails", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          errors: [{ detail: "Override cannot be removed yet" }],
        }),
        {
          status: 409,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );
    const { setOverrides } = renderComponent();

    await user.click(
      screen.getAllByRole("button", { name: "Remove override" })[0],
    );

    expect(setOverrides).not.toHaveBeenCalled();
    expect(screen.getByText("Override cannot be removed yet")).toBeTruthy();
  });

  it("falls back to the HTTP status when the error response has no detail", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response("{}", {
        status: 500,
        headers: { "Content-Type": "application/json" },
      }),
    );

    renderComponent();
    await user.click(
      screen.getAllByRole("button", { name: "Remove override" })[0],
    );

    expect(screen.getByText("HTTP 500")).toBeTruthy();
  });

  it("renders an unexpected error message when fetch throws", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("network down"));
    const { setOverrides } = renderComponent();

    await user.click(
      screen.getAllByRole("button", { name: "Remove override" })[0],
    );

    expect(setOverrides).not.toHaveBeenCalled();
    expect(screen.getByText("An unexpected error occurred")).toBeTruthy();
  });

  it("renders an unexpected error when the rotation id is unavailable", async () => {
    const user = userEvent.setup();
    const fetchMock = vi.spyOn(globalThis, "fetch");
    useParamsMock.mockReturnValue({});
    const { setOverrides } = renderComponent();

    await user.click(
      screen.getAllByRole("button", { name: "Remove override" })[0],
    );

    expect(fetchMock).not.toHaveBeenCalled();
    expect(setOverrides).not.toHaveBeenCalled();
    expect(screen.getByText("An unexpected error occurred")).toBeTruthy();
  });

  it("disables all delete buttons while the delete request is in flight", async () => {
    const user = userEvent.setup();
    const deferred = createDeferredResponse();
    vi.spyOn(globalThis, "fetch").mockReturnValue(deferred.promise);

    renderComponent();
    const buttons = screen.getAllByRole("button", { name: "Remove override" });

    const clickPromise = user.click(buttons[0]);

    await waitFor(() => {
      for (const button of screen.getAllByRole("button", {
        name: "Remove override",
      })) {
        expect(button).toHaveProperty("disabled", true);
      }
    });

    deferred.resolve(new Response(null, { status: 204 }));
    await clickPromise;

    await waitFor(() => {
      for (const button of screen.getAllByRole("button", {
        name: "Remove override",
      })) {
        expect(button).toHaveProperty("disabled", false);
      }
    });
  });
});
