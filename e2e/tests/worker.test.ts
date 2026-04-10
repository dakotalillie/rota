import * as path from "path";
import { expect, test } from "../fixtures";

test.describe("worker", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-handoff-pending.json") });

  test("worker advances rotation at handoff time", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    // Set time to just before the handoff (Monday 2026-04-06 08:30 PDT = 15:30 UTC)
    setTime("2026-04-06T15:30:00Z");

    await page.goto(`${serverUrl}/rotations/rot_01SEED000000000000000ROT1`);

    // Alice should be on-call before the handoff
    await expect(page.getByRole("heading", { name: "Alice Smith", level: 2 })).toBeVisible();

    // Advance time to after the handoff (Monday 2026-04-06 09:30 PDT = 16:30 UTC)
    setTime("2026-04-06T16:30:00Z");

    // Wait for the worker to tick (worker interval is 5s, wait 8s to be safe)
    await page.waitForTimeout(8000);

    // Reload to get fresh data
    await page.reload();

    // Bob should now be on-call
    await expect(page.getByRole("heading", { name: "Bob Jones", level: 2 })).toBeVisible();
  });
});
