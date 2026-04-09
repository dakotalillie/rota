import { test as base } from "@playwright/test";
import { ChildProcess, spawn } from "child_process";
import * as fs from "fs";
import * as net from "net";
import * as os from "os";
import * as path from "path";

const SERVER_BINARY = path.resolve(__dirname, "bin/server");
const STATIC_DIR = path.resolve(__dirname, "../ui/dist");

function getFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      if (!address || typeof address === "string") {
        reject(new Error("Failed to get free port"));
        return;
      }
      const port = address.port;
      server.close(() => resolve(port));
    });
    server.on("error", reject);
  });
}

async function waitForServer(url: string, timeout = 10_000): Promise<void> {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(url);
      if (res.ok || res.status === 404) return;
    } catch {
      // not ready yet
    }
    await new Promise((r) => setTimeout(r, 100));
  }
  throw new Error(`Server at ${url} did not become ready within ${timeout}ms`);
}

type ServerFixture = {
  serverUrl: string;
};

export const test = base.extend<ServerFixture>({
  serverUrl: async ({}, use) => {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "rota-e2e-"));
    const dbPath = path.join(tmpDir, "rota.db");
    const port = await getFreePort();
    const baseUrl = `http://localhost:${port}`;

    const proc: ChildProcess = spawn(SERVER_BINARY, [], {
      env: {
        ...process.env,
        DATABASE_PATH: dbPath,
        PORT: String(port),
        HOSTNAME: baseUrl,
        STATIC_DIR: STATIC_DIR,
      },
      stdio: "pipe",
    });

    proc.stderr?.pipe(process.stderr);

    await waitForServer(`${baseUrl}/api/rotations`);

    await use(baseUrl);

    proc.kill("SIGTERM");
    await new Promise<void>((resolve) => proc.on("close", () => resolve()));
    fs.rmSync(tmpDir, { recursive: true, force: true });
  },

  page: async ({ page, serverUrl }, use) => {
    await page.goto(serverUrl);
    await use(page);
  },
});

export { expect } from "@playwright/test";
