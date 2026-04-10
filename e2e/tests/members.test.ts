import * as path from "path";
import { expect, test } from "../fixtures";

test.describe("adding members to an empty rotation", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-empty.json") });

  test("add a member and verify they appear", async ({ page, serverUrl }) => {
    await page.goto(`${serverUrl}/rotations`);
    await page.getByText("Backend On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    await page.getByRole("button", { name: "Add member" }).click();
    await page.getByPlaceholder("Name").fill("Alice Smith");
    await page.getByPlaceholder("Email").fill("alice@example.com");
    await page.getByRole("button", { name: "Add person" }).click();

    const membersList = page.getByTestId("members-list");
    await expect(membersList.getByTestId("member-row")).toHaveText([
      /Alice Smith/,
    ]);
  });

  test("add multiple members and verify order", async ({ page, serverUrl }) => {
    await page.goto(`${serverUrl}/rotations`);
    await page.getByText("Backend On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    const members = [
      { name: "Alice Smith", email: "alice@example.com" },
      { name: "Bob Jones", email: "bob@example.com" },
      { name: "Carol White", email: "carol@example.com" },
    ];

    for (const m of members) {
      await page.getByRole("button", { name: "Add member" }).click();
      await page.getByPlaceholder("Name").fill(m.name);
      await page.getByPlaceholder("Email").fill(m.email);
      await page.getByRole("button", { name: "Add person" }).click();
    }

    const memberRows = page.getByTestId("members-list").getByTestId("member-row");
    await expect(memberRows).toHaveText([
      /Alice Smith/,
      /Bob Jones/,
      /Carol White/,
    ]);
  });
});

test.describe("with seeded members", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-with-members.json") });

  test("delete a non-current member removes them from the list", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    const membersList = page.getByTestId("members-list");
    await page.getByRole("button", { name: "Remove Carol White" }).click();

    await expect(membersList.getByText("Carol White")).toHaveCount(0);
    await expect(membersList.getByText("Alice Smith")).toBeVisible();
  });

  test("delete current on-call member promotes next", async ({
    page,
    serverUrl,
    setTime,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);

    await page.getByText("Platform On-Call").click();
    await page.getByRole("button", { name: "Settings" }).click();

    const membersList = page.getByTestId("members-list");
    await page.getByRole("button", { name: "Remove Alice Smith" }).click();

    await expect(membersList.getByText("Bob Jones")).toBeVisible();
    await expect(membersList.getByText("Alice Smith")).toHaveCount(0);
  });
});
