// @vitest-environment jsdom
// The comment above tells Vitest to run this file in jsdom, a fake browser,
// so components can be drawn without a real one. Learn more: docs/learn/testing.md
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { NewestGreeting } from "./NewestGreeting.tsx";

// Remove what each test drew, so the next test starts with an empty page.
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

test("shows the greeting once it has loaded", async () => {
  const loadGreeting = async () => ({ id: 2n, message: "Hello, test!" });

  render(<NewestGreeting loadGreeting={loadGreeting} />);

  // getByText fails the test if the text is not on the page right now.
  screen.getByText("Loading the greeting…");
  // findByText waits (up to 1 second) for the text to appear.
  await screen.findByText("Hello, test!");
  await screen.findByText("Greeting #2");
});

test("shows an error when loading fails", async () => {
  const loadGreeting = async () => {
    throw new Error("network down");
  };
  // Replace console.error with a silent stand-in, so the test output stays
  // clean, and so we can check that the error was logged.
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  render(<NewestGreeting loadGreeting={loadGreeting} />);

  await screen.findByText("Could not load the greeting.");
  expect(consoleError).toHaveBeenCalled();
});
