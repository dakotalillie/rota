// @vitest-environment jsdom

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import Members from "./Members";
import type { Member, Override } from "./types";

vi.mock("@tanstack/react-router", () => ({
  useParams: () => ({ rotationId: "rot_123" }),
}));

vi.mock("./AddMemberDialog", () => ({
  default: () => null,
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

describe("Members", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  function renderComponent(setMembers = vi.fn(), setOverrides = vi.fn()) {
    render(
      <Members
        members={members}
        setMembers={setMembers}
        overrides={overrides}
        setOverrides={setOverrides}
      />,
    );

    return { setMembers, setOverrides };
  }

  it("deletes a member and removes related overrides after a successful response", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response(null, { status: 204 }));
    const { setMembers, setOverrides } = renderComponent();

    await user.click(screen.getByRole("button", { name: "Remove Bob Brown" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/rotations/rot_123/members/mem_2",
      { method: "DELETE" },
    );
    expect(setMembers).toHaveBeenCalledWith([members[0]]);
    expect(setOverrides).toHaveBeenCalledWith([]);
  });

  it("renders the server error detail and preserves state when deletion fails", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          errors: [{ detail: "Member cannot be removed yet" }],
        }),
        {
          status: 409,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );
    const { setMembers, setOverrides } = renderComponent();

    await user.click(screen.getByRole("button", { name: "Remove Bob Brown" }));

    expect(setMembers).not.toHaveBeenCalled();
    expect(setOverrides).not.toHaveBeenCalled();
    expect(screen.getByText("Member cannot be removed yet")).toBeTruthy();
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
    await user.click(screen.getByRole("button", { name: "Remove Bob Brown" }));

    expect(screen.getByText("HTTP 500")).toBeTruthy();
  });

  it("renders an unexpected error message when fetch throws", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("network down"));
    const { setMembers, setOverrides } = renderComponent();

    await user.click(screen.getByRole("button", { name: "Remove Bob Brown" }));

    expect(setMembers).not.toHaveBeenCalled();
    expect(setOverrides).not.toHaveBeenCalled();
    expect(screen.getByText("An unexpected error occurred")).toBeTruthy();
  });

  it("disables the clicked button while the delete request is in flight", async () => {
    const user = userEvent.setup();
    const deferred = createDeferredResponse();
    vi.spyOn(globalThis, "fetch").mockReturnValue(deferred.promise);

    renderComponent();
    const button = screen.getByRole("button", { name: "Remove Bob Brown" });

    const clickPromise = user.click(button);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Remove Bob Brown" }),
      ).toHaveProperty("disabled", true);
    });

    deferred.resolve(new Response(null, { status: 204 }));
    await clickPromise;

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Remove Bob Brown" }),
      ).toHaveProperty("disabled", false);
    });
  });
});
