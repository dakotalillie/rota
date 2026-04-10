import * as path from "path";
import { expect, test } from "../fixtures";

test.describe("overrides", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-with-members.json") });

  test("create an override and verify it appears", async ({
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
    await dialog.getByLabel("Override start").fill("2026-04-08T10:00");
    await dialog.getByLabel("Override end").fill("2026-04-09T10:00");
    await dialog.getByLabel("Override member").selectOption({ label: "Bob Jones" });
    await dialog.getByRole("button", { name: "Add override" }).click();

    await expect(page.getByRole("button", { name: "Remove override" })).toBeVisible();
  });
});

test.describe("override changes effective on-call display", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-with-override.json") });

  test("shows override member as on-call when time is within override window", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();

    await expect(page.getByRole("heading", { name: "Bob Jones", level: 2 })).toBeVisible();
  });

  test("rotation list shows override member as on-call", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await expect(page.getByText("Bob Jones")).toBeVisible();
  });

  test("delete an override removes it from the list", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    await page.getByRole("button", { name: "Remove override" }).click();

    await expect(page.getByText("No overrides scheduled.")).toBeVisible();
  });
});
