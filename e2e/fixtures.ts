import { test as base } from "@playwright/test";
import { ChildProcess, spawn, spawnSync } from "child_process";
import * as fs from "fs";
import * as net from "net";
import * as os from "os";
import * as path from "path";

const SERVER_BINARY = path.resolve(__dirname, "bin/server");
const SEED_BINARY = path.resolve(__dirname, "bin/seed");
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

type Ctx = {
  tmpDir: string;
  dbPath: string;
  timeOverrideFile: string;
};

type Fixtures = {
  _ctx: Ctx;
  seedFile: string | undefined;
  serverUrl: string;
  setTime: (isoString: string) => void;
  clearTime: () => void;
  api: (method: string, urlPath: string, body?: unknown) => Promise<Response>;
};

export const test = base.extend<Fixtures>({
  _ctx: async ({}, use) => {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "rota-e2e-"));
    const dbPath = path.join(tmpDir, "rota.db");
    const timeOverrideFile = path.join(tmpDir, "time-override.txt");
    await use({ tmpDir, dbPath, timeOverrideFile });
    fs.rmSync(tmpDir, { recursive: true, force: true });
  },

  seedFile: [undefined, { option: true }],

  serverUrl: async ({ _ctx, seedFile }, use) => {
    const { dbPath, timeOverrideFile } = _ctx;
    const port = await getFreePort();
    const baseUrl = `http://localhost:${port}`;

    if (seedFile) {
      const result = spawnSync(
        SEED_BINARY,
        [`-db=${dbPath}`, `-seed-file=${seedFile}`],
        {
          encoding: "utf8",
        },
      );
      if (result.status !== 0) {
        throw new Error(`Seed binary failed:\n${result.stderr}`);
      }
    }

    const proc: ChildProcess = spawn(SERVER_BINARY, [], {
      env: {
        ...process.env,
        DATABASE_PATH: dbPath,
        PORT: String(port),
        HOSTNAME: baseUrl,
        STATIC_DIR: STATIC_DIR,
        TIME_OVERRIDE_FILE: timeOverrideFile,
      },
      stdio: "pipe",
    });

    const showLogs = process.env.E2E_SERVER_LOGS === "1";
    let stderrBuffer = "";
    proc.stderr?.on("data", (chunk: Buffer) => {
      if (showLogs) {
        process.stderr.write(chunk);
      } else {
        stderrBuffer += chunk.toString();
      }
    });

    try {
      await waitForServer(`${baseUrl}/api/rotations`);
    } catch (err) {
      if (stderrBuffer) {
        process.stderr.write(stderrBuffer);
      }
      throw err;
    }

    await use(baseUrl);

    proc.kill("SIGTERM");
    await new Promise<void>((resolve) => proc.on("close", () => resolve()));
  },

  setTime: async ({ _ctx }, use) => {
    await use((isoString: string) => {
      fs.writeFileSync(_ctx.timeOverrideFile, isoString);
    });
  },

  clearTime: async ({ _ctx }, use) => {
    await use(() => {
      try {
        fs.unlinkSync(_ctx.timeOverrideFile);
      } catch {
        // ignore if file doesn't exist
      }
    });
  },

  api: async ({ serverUrl }, use) => {
    await use(async (method: string, urlPath: string, body?: unknown) => {
      const url = `${serverUrl}${urlPath}`;
      const init: RequestInit = {
        method,
        headers: { "Content-Type": "application/json" },
      };
      if (body !== undefined) {
        init.body = JSON.stringify(body);
      }
      return fetch(url, init);
    });
  },

  page: async ({ page, serverUrl }, use) => {
    await page.goto(serverUrl);
    await use(page);
  },
});

export { expect } from "@playwright/test";
