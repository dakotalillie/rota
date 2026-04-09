import { expect, test } from "../fixtures";

test("shows empty state when no rotations exist", async ({
  page,
  serverUrl,
}) => {
  await page.goto(`${serverUrl}/rotations`);
  await expect(page.getByText("No rotations yet.")).toBeVisible();
});

test("creates a rotation and shows it in the list", async ({
  page,
  serverUrl,
}) => {
  await page.goto(`${serverUrl}/rotations`);

  await page.getByRole("button", { name: "Create Rotation" }).click();
  await page.getByPlaceholder("Rotation name").fill("Backend On-Call");
  await page.getByRole("button", { name: "Create" }).click();

  await expect(page.getByText("Backend On-Call")).toBeVisible();
});

test("navigates to rotation detail after clicking a rotation", async ({
  page,
  serverUrl,
}) => {
  await page.goto(`${serverUrl}/rotations`);

  await page.getByRole("button", { name: "Create Rotation" }).click();
  await page.getByPlaceholder("Rotation name").fill("Infra On-Call");
  await page.getByRole("button", { name: "Create" }).click();

  await page.getByText("Infra On-Call").click();

  await expect(page).toHaveURL(/\/rotations\/.+/);
});
