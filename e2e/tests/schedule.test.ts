import * as path from "path";

import { expect, test } from "../fixtures";

test.describe("schedule", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-members.json"),
  });

  test("schedule shows members in rotation order", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);
    await page.getByText("Platform On-Call").click();

    const rows = page.getByTestId("schedule").getByTestId("schedule-row");
    await expect(rows.nth(0)).toContainText("Alice Smith");
    await expect(rows.nth(1)).toContainText("Bob Jones");
    await expect(rows.nth(2)).toContainText("Carol White");
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
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);
    await page.getByText("Platform On-Call").click();

    const schedule = page.getByTestId("schedule");
    const overrideRow = schedule
      .getByTestId("schedule-row")
      .filter({ hasText: "Override" });
    await expect(overrideRow).toContainText("Bob Jones");
  });
});
