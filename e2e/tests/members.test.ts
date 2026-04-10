import * as path from "path";
import { expect, test } from "../fixtures";

test("add a member and verify they appear", async ({ page, serverUrl }) => {
  await page.goto(`${serverUrl}/rotations`);

  await page.getByRole("button", { name: "Create Rotation" }).click();
  await page.getByPlaceholder("Rotation name").fill("Backend On-Call");
  await page.getByRole("button", { name: "Create" }).click();

  await page.getByText("Backend On-Call").click();
  await page.getByRole("button", { name: "Settings" }).click();

  await page.getByRole("button", { name: "Add member" }).click();
  await page.getByPlaceholder("Name").fill("Alice Smith");
  await page.getByPlaceholder("Email").fill("alice@example.com");
  await page.getByRole("button", { name: "Add person" }).click();

  await expect(page.getByText("Alice Smith")).toBeVisible();
});

test("add multiple members and verify order", async ({ page, serverUrl }) => {
  await page.goto(`${serverUrl}/rotations`);

  await page.getByRole("button", { name: "Create Rotation" }).click();
  await page.getByPlaceholder("Rotation name").fill("Backend On-Call");
  await page.getByRole("button", { name: "Create" }).click();

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
    await expect(page.getByText(m.name)).toBeVisible();
  }

  const names = await page.getByText(/Smith|Jones|White/).allTextContents();
  expect(names[0]).toContain("Alice");
  expect(names[1]).toContain("Bob");
  expect(names[2]).toContain("Carol");
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

    await page.getByRole("button", { name: "Remove Carol White" }).click();

    await expect(page.getByText("Carol White")).not.toBeVisible();
    await expect(page.getByText("Alice Smith")).toBeVisible();
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

    await page.getByRole("button", { name: "Remove Alice Smith" }).click();

    await expect(page.getByText("Bob Jones")).toBeVisible();
    await expect(page.getByText("Alice Smith")).not.toBeVisible();
  });
});
