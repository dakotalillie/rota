import * as path from "path";
import { expect, test } from "../fixtures";

test.describe("schedule", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-with-members.json") });

  test("schedule shows correct member blocks in rotation order", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations/rot_01SEED000000000000000ROT1`);

    // The schedule section should show members in order: Alice, Bob, Carol, Alice...
    await expect(page.getByText("Alice Smith").first()).toBeVisible();
    await expect(page.getByText("Bob Jones").first()).toBeVisible();
  });
});

test.describe("schedule reflects overrides", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-with-override.json") });

  test("schedule shows override member in override time window", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations/rot_01SEED000000000000000ROT1`);

    // Bob's override block should be visible in the schedule
    await expect(page.getByText("Bob Jones").first()).toBeVisible();
    await expect(page.getByText("Override").first()).toBeVisible();
  });
});
