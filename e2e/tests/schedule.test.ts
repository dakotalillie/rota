import * as path from "path";

import { expect, test } from "../fixtures";
import { sanitizeApiResponse } from "../snapshot-utils";

const SEED_ROT_ID = "rot_01SEED000000000000000ROT1";

test.describe("schedule", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-members.json"),
  });

  test("schedule shows members in rotation order", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);
    await page.getByText("Platform On-Call").click();

    const rows = page.getByTestId("schedule").getByTestId("schedule-row");
    await expect(rows.nth(0)).toContainText("Alice Smith");
    await expect(rows.nth(1)).toContainText("Bob Jones");
    await expect(rows.nth(2)).toContainText("Carol White");

    const res = await api(
      "GET",
      `/api/rotations/${SEED_ROT_ID}/schedule?weeks=4`,
    );
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "members-in-rotation-order.json",
    );
  });
});

test.describe("schedule reflects overrides", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-override.json"),
  });

  test("schedule shows override member in override time window", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);
    await page.getByText("Platform On-Call").click();

    const schedule = page.getByTestId("schedule");
    const overrideRow = schedule
      .getByTestId("schedule-row")
      .filter({ hasText: "Override" });
    await expect(overrideRow).toContainText("Bob Jones");

    const res = await api(
      "GET",
      `/api/rotations/${SEED_ROT_ID}/schedule?weeks=4`,
    );
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "override-in-schedule.json",
    );
  });
});
