import * as path from "path";

import { expect, test } from "../fixtures";
import { sanitizeApiResponse } from "../snapshot-utils";

const SEED_ROT_ID = "rot_01SEED000000000000000ROT1";

test.describe("overrides", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-members.json"),
    timezoneId: "America/Los_Angeles",
  });

  test("previews correct shift boundaries at the cadence handoff time", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    await page.getByRole("button", { name: "Add override" }).click();

    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("Override start").fill("2026-04-15T10:00");
    await dialog.getByLabel("Override end").fill("2026-04-20T14:00");
    await dialog
      .getByLabel("Override member")
      .selectOption({ label: "Alice Smith" });

    const replacesPanel = dialog.locator("text=Replaces").locator("..");
    await expect(replacesPanel).toContainText(
      "Apr 15, 10:00 AM – Apr 20, 9:00 AM",
    );
    await expect(replacesPanel).toContainText(
      "Apr 20, 9:00 AM – Apr 20, 2:00 PM",
    );
    await expect(replacesPanel).not.toContainText("12:00 AM");
  });

  test("preview excludes adjacent block when override ends exactly on a boundary", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    await page.getByRole("button", { name: "Add override" }).click();

    const dialog = page.getByRole("dialog");
    // Override ends exactly at the Apr 20 9am boundary between Bob and Carol's blocks
    await dialog.getByLabel("Override start").fill("2026-04-17T09:00");
    await dialog.getByLabel("Override end").fill("2026-04-20T09:00");
    await dialog
      .getByLabel("Override member")
      .selectOption({ label: "Alice Smith" });

    const replacesPanel = dialog.locator("text=Replaces").locator("..");
    // Only Bob's clipped segment should appear
    await expect(replacesPanel).toContainText("Bob Jones");
    await expect(replacesPanel).toContainText(
      "Apr 17, 9:00 AM – Apr 20, 9:00 AM",
    );
    // Carol's block starts exactly at the override end — it must not appear
    await expect(replacesPanel).not.toContainText("Carol White");
  });

  test("create an override and verify it appears", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    await page.getByRole("button", { name: "Add override" }).click();

    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("Override start").fill("2026-04-08T10:00");
    await dialog.getByLabel("Override end").fill("2026-04-09T10:00");
    await dialog
      .getByLabel("Override member")
      .selectOption({ label: "Bob Jones" });
    await dialog.getByRole("button", { name: "Add override" }).click();

    await expect(
      page.getByRole("button", { name: "Remove override" }),
    ).toBeVisible();

    const res = await api("GET", `/api/rotations/${SEED_ROT_ID}`);
    const body = await res.json();
    expect(sanitizeApiResponse(body, { maskNewIds: true })).toMatchSnapshot(
      "create-override.json",
    );
  });
});

test.describe("override changes effective on-call display", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-override.json"),
  });

  test("shows override member as on-call when time is within override window", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();

    await expect(
      page.getByRole("heading", { name: "Bob Jones", level: 2 }),
    ).toBeVisible();

    const res = await api("GET", `/api/rotations/${SEED_ROT_ID}`);
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "shows-override-as-on-call.json",
    );
  });

  test("rotation list shows override member as on-call", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await expect(page.getByText("Bob Jones")).toBeVisible();

    const res = await api("GET", "/api/rotations");
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "list-shows-override-on-call.json",
    );
  });

  test("delete an override removes it from the list", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    await page.getByRole("button", { name: "Remove override" }).click();

    await expect(page.getByText("No overrides scheduled.")).toBeVisible();

    const res = await api("GET", `/api/rotations/${SEED_ROT_ID}`);
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot("delete-override.json");
  });
});
