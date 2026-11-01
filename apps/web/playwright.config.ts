// Settings for Playwright, which runs the end-to-end tests in e2e/ in a
// real browser. Run them with "task test:e2e" from the repo's top folder:
// it starts the system first. Learn more: docs/learn/playwright.md
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  // Our test files end in .e2e.ts. (Vitest, for the unit tests, looks for
  // .test.ts files, so the two never pick up each other's files.)
  testMatch: "**/*.e2e.ts",
  // In CI, fail if a test was left with "test.only" (which skips all the
  // others). The CI environment variable is set by GitHub Actions.
  forbidOnly: process.env.CI !== undefined,
  use: {
    // The gateway of the test system (e2e/compose.e2e.yaml). page.goto("/")
    // opens this address.
    baseURL: "http://localhost:8180",
    // When a test fails, keep a "trace": a recording of every step, with
    // screenshots, to look at with "npx playwright show-trace".
    trace: "retain-on-failure",
  },
  // One browser engine for now (ADR 0034). Firefox and WebKit (Safari's
  // engine) are one line each, later.
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
