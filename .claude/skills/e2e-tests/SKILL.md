---
name: e2e-tests
description: Write or modify Playwright E2E tests for the rota project. Use when adding, editing, or debugging tests under e2e/, or when the user asks to test UI behavior end-to-end. Covers the custom fixtures (serverUrl, seedFile, setTime, api), seed-file workflow, and Playwright selector/assertion conventions.
paths: e2e/**
---

# Rota E2E test playbook

This skill applies whenever you're adding, editing, or debugging a Playwright test under `e2e/tests/`, touching `e2e/fixtures.ts`, or creating a seed file under `e2e/seed/`.

## Project conventions (non-negotiable)

### Import from the custom fixtures file

```ts
import { expect, test } from "../fixtures";
```

**Never** `import { test } from "@playwright/test"` — the custom `test` in `e2e/fixtures.ts` extends the base with project fixtures. `expect` is re-exported from the same file for convenience.

### Available fixtures (see `e2e/fixtures.ts`)

- **`serverUrl`** — a fresh Go server with an isolated temp SQLite DB per test. Always navigate with `` await page.goto(`${serverUrl}/some/path`) ``. The built-in `page` fixture auto-navigates to `serverUrl` (root), but tests typically need a specific route, so an explicit `page.goto` is still the norm.
- **`seedFile`** — set per-describe-block via `test.use({ seedFile: path.join(__dirname, "../seed/<file>.json") })`. The seed binary runs before the server starts. Without a seed file the DB is empty.
- **`setTime(iso)`** / **`clearTime()`** — writes to a `TIME_OVERRIDE_FILE` the server reads, so the Go clock returns your chosen instant. Use for anything schedule- or handoff-sensitive.
- **`api(method, urlPath, body?)`** — thin `fetch` wrapper for hitting the JSON:API directly. Use when UI setup would be noisy or when asserting on API behavior.

### Seed files

Located in `e2e/seed/`. Before creating a new one, check whether an existing file covers your starting state:

- `rotation-with-members.json` — one rotation with a few members; default choice for tests that need members present.
- `rotation-with-override.json` — one rotation with an active override.
- `rotation-handoff-pending.json` — rotation positioned right before a handoff; used by worker/schedule tests that need deterministic IDs and timing.

Only add a new seed file when no existing file fits. Keep seed data minimal.

### Running tests

Use `task test:e2e` (per the repo convention of preferring Taskfile commands — see AGENTS.md). `task build:e2e` builds the Go `server` and `seed` binaries the fixtures need; `task install:e2e` installs Playwright. Don't invoke `npx playwright test` directly as the standard run.

For local debugging only, these direct commands are fine:
- `npx playwright show-trace <path>` — inspect a saved trace
- `npx playwright test --debug` — step through a test

## Playwright best practices

### Locators — prefer user-facing roles

```ts
page.getByRole("button", { name: "Create Rotation" })
page.getByRole("heading", { name: "Alice Smith", level: 2 })
page.getByPlaceholder("Rotation name")
page.getByLabel("Email")
page.getByText("No rotations yet.")
page.getByTestId("rotation-row")  // last resort when nothing else is semantic
```

Filter instead of writing complex selectors:

```ts
page.getByRole("listitem").filter({ hasText: "Alice" })
```

Avoid CSS/XPath (`page.locator(".foo")`, `page.locator("//div")`) — they break on DOM changes and ignore accessibility.

### Assertions — web-first, always

```ts
await expect(page.getByText("Backend On-Call")).toBeVisible();
await expect(page).toHaveURL(/\/rotations\/.+/);
await expect(page.getByRole("listitem")).toHaveCount(3);
```

Web-first assertions auto-retry until the condition is met (or timeout). Do **not** write:

```ts
// ❌ skips auto-waiting, flaky
expect(await locator.isVisible()).toBe(true);
expect(await locator.textContent()).toBe("Alice");
```

### No manual waits (with one exception)

Don't use `page.waitForTimeout` to "give the UI a moment." Locators auto-wait for actionability, and assertions retry. If a test feels like it needs a sleep, the fix is almost always an assertion on the thing you were waiting for.

**The one legitimate exception** is waiting for wall-clock driven background work on the Go side — e.g. the rotation worker that ticks on a fixed interval. `e2e/tests/worker.test.ts` uses `page.waitForTimeout(8000)` for exactly this reason. If you add another case like this, leave a comment explaining which interval you're waiting on.

### Test isolation

Each test gets a fresh server and DB via `serverUrl`. Never rely on ordering or leaked state between tests. Parallel execution is enabled.

### Test names

Use narrative names describing user-observable behavior: `"creates a rotation and shows it in the list"`, not `"test_create"`.

## Anti-patterns (don't do these)

- `import { test } from "@playwright/test"` — loses project fixtures
- `page.locator(".some-class")` / XPath — brittle, use role-based locators
- `await page.waitForTimeout(N)` — use assertions instead (see exception above)
- `expect(await locator.textContent()).toBe(...)` — use `toHaveText`, which auto-retries
- Creating test data through the UI when a seed file already covers it
- Introducing a Page Object Model — this codebase uses the Playwright API directly; stay consistent

## Workflow for a new test

1. **Seed:** does an existing file in `e2e/seed/` cover the starting state? Reuse it. Only create a new one if nothing fits.
2. **File:** pick the existing test file that matches the feature area (`rotations.test.ts`, `members.test.ts`, `overrides.test.ts`, `schedule.test.ts`, `worker.test.ts`). Only create a new file for a genuinely new feature area.
3. **Shape:** if multiple tests share a seed, group them in a `test.describe` with `test.use({ seedFile: ... })` at the top.
4. **Time:** if the test is schedule- or handoff-sensitive, call `setTime("...")` before the first `page.goto`.
5. **Write** using fixtures from `../fixtures`, role-based locators, web-first assertions.
6. **Run** with `task test:e2e`. On failure, re-run and inspect traces from `e2e/playwright-report/` via `npx playwright show-trace`.

## Canonical examples to read before writing

- `e2e/fixtures.ts` — fixture definitions (read this first if you're doing anything non-obvious)
- `e2e/tests/rotations.test.ts` — plain test style and a simple `describe` + `seedFile` block
- `e2e/tests/worker.test.ts` — `setTime` usage and the rare `waitForTimeout` exception
- `e2e/tests/overrides.test.ts` — seeded state + UI interaction
- `e2e/playwright.config.ts` — config (rarely needs to change)
