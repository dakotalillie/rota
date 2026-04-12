import * as path from "path";

import { expect, test } from "../fixtures";
import { sanitizeApiResponse } from "../snapshot-utils";

test("shows empty state when no rotations exist", async ({
  page,
  serverUrl,
  api,
}) => {
  await page.goto(`${serverUrl}/rotations`);
  await expect(page.getByText("No rotations yet.")).toBeVisible();

  const res = await api("GET", "/api/rotations");
  const body = await res.json();
  expect(sanitizeApiResponse(body)).toMatchSnapshot("empty-state.json");
});

test("creates a rotation and shows it in the list", async ({
  page,
  serverUrl,
  api,
}) => {
  await page.goto(`${serverUrl}/rotations`);

  await page.getByRole("button", { name: "Create Rotation" }).click();
  await page.getByPlaceholder("Rotation name").fill("Backend On-Call");
  await page.getByRole("button", { name: "Create" }).click();

  await expect(page.getByText("Backend On-Call")).toBeVisible();

  const res = await api("GET", "/api/rotations");
  const body = await res.json();
  expect(sanitizeApiResponse(body, { maskNewIds: true })).toMatchSnapshot(
    "creates-rotation.json",
  );
});

test("navigates to rotation detail after clicking a rotation", async ({
  page,
  serverUrl,
  api,
}) => {
  await page.goto(`${serverUrl}/rotations`);

  await page.getByRole("button", { name: "Create Rotation" }).click();
  await page.getByPlaceholder("Rotation name").fill("Infra On-Call");
  await page.getByRole("button", { name: "Create" }).click();

  await page.getByText("Infra On-Call").click();

  await expect(page).toHaveURL(/\/rotations\/.+/);

  const rotationId = page.url().match(/\/rotations\/([^/]+)/)?.[1];
  const res = await api("GET", `/api/rotations/${rotationId}`);
  const body = await res.json();
  expect(sanitizeApiResponse(body, { maskNewIds: true })).toMatchSnapshot(
    "navigates-to-detail.json",
  );
});

test.describe("deleting a rotation", () => {
  test.use({ seedFile: path.join(__dirname, "../seed/rotation-empty.json") });

  test("delete a rotation removes it from the list", async ({
    page,
    serverUrl,
    api,
  }) => {
    await page.goto(`${serverUrl}/rotations`);

    await page.getByRole("button", { name: "Delete Backend On-Call" }).click();
    await page.getByRole("button", { name: "Delete" }).click();

    await expect(page.getByText("Backend On-Call")).toHaveCount(0);
    await expect(page.getByText("No rotations yet.")).toBeVisible();

    const res = await api("GET", "/api/rotations");
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "delete-removes-from-list.json",
    );
  });

  test("cancel does not delete the rotation", async ({
    page,
    serverUrl,
    api,
  }) => {
    await page.goto(`${serverUrl}/rotations`);

    await page.getByRole("button", { name: "Delete Backend On-Call" }).click();
    await page.getByRole("button", { name: "Cancel" }).click();

    await expect(page.getByText("Backend On-Call")).toBeVisible();

    const res = await api("GET", "/api/rotations");
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "cancel-does-not-delete.json",
    );
  });
});

test.describe("rotation list shows current on-call member", () => {
  test.use({
    seedFile: path.join(__dirname, "../seed/rotation-with-members.json"),
  });

  test("shows the current on-call member name in the list", async ({
    page,
    serverUrl,
    setTime,
    api,
  }) => {
    setTime("2026-04-07T12:00:00Z");
    await page.goto(`${serverUrl}/rotations`);
    await expect(page.getByText("Alice Smith")).toBeVisible();

    const res = await api("GET", "/api/rotations");
    const body = await res.json();
    expect(sanitizeApiResponse(body)).toMatchSnapshot(
      "shows-current-member.json",
    );
  });
});
