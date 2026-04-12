import * as path from "path";

import { expect, test } from "../fixtures";
import { sanitizeApiResponse } from "../snapshot-utils";

const SEED_ROT_ID = "rot_01SEED000000000000000ROT1";

test.describe("overrides", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-members.json"),
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
