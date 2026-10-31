// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ConnectError, Code } from "@connectrpc/connect";
import { afterEach, expect, test, vi } from "vitest";
import type { AuthApi, Device } from "./authClient.ts";
import { DeviceList } from "./DeviceList.tsx";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

const morning = new Date("2026-10-31T09:00:00Z");

const laptop: Device = {
  id: "019a3f2c-0000-7000-8000-000000000001",
  name: "Chrome on Windows",
  loggedInAt: morning,
  lastUsedAt: morning,
  isCurrentDevice: true,
};

const phone: Device = {
  id: "019a3f2c-0000-7000-8000-000000000002",
  name: "Safari on iPhone",
  loggedInAt: morning,
  lastUsedAt: morning,
  isCurrentDevice: false,
};

// fakeAuthApi returns an AuthApi with these two devices. Only the device
// calls matter here; the others fail if a test uses them by mistake.
function fakeAuthApi(changes: Partial<AuthApi> = {}): AuthApi {
  const notUsedHere = async () => {
    throw new Error("not used in this test");
  };
  const defaults: AuthApi = {
    signUp: notUsedHere,
    logIn: notUsedHere,
    getCurrentUser: notUsedHere,
    logOut: notUsedHere,
    listDevices: async () => [laptop, phone],
    logOutDevice: async () => {},
  };
  return { ...defaults, ...changes };
}

test("lists the devices and marks this one", async () => {
  render(<DeviceList authApi={fakeAuthApi()} />);

  await screen.findByText("Chrome on Windows");
  screen.getByText("Safari on iPhone");
  screen.getByText("This device");
  // Only the other device gets a log-out button.
  screen.getByRole("button", { name: "Log out Safari on iPhone" });
  expect(
    screen.queryByRole("button", { name: "Log out Chrome on Windows" }),
  ).toBeNull();
});

test("logs out another device and removes it from the list", async () => {
  const logOutDevice = vi.fn(async () => {});
  render(<DeviceList authApi={fakeAuthApi({ logOutDevice })} />);
  await screen.findByText("Safari on iPhone");

  fireEvent.click(
    screen.getByRole("button", { name: "Log out Safari on iPhone" }),
  );

  // The click's work finishes a moment later. vi.waitFor repeats a check
  // until it passes, or fails the test after one second.
  await vi.waitFor(() => expect(logOutDevice).toHaveBeenCalledWith(phone.id));
  await vi.waitFor(() =>
    expect(screen.queryByText("Safari on iPhone")).toBeNull(),
  );
  screen.getByText("Chrome on Windows");
});

test("shows the message when logging out a device fails", async () => {
  const logOutDevice = async () => {
    throw new ConnectError("you have no device with this id", Code.NotFound);
  };
  render(<DeviceList authApi={fakeAuthApi({ logOutDevice })} />);
  await screen.findByText("Safari on iPhone");

  fireEvent.click(
    screen.getByRole("button", { name: "Log out Safari on iPhone" }),
  );

  await screen.findByText("You have no device with this id");
  // The device stays in the list, because it was not logged out.
  screen.getByText("Safari on iPhone");
});

test("shows an error when the devices can't be loaded", async () => {
  const listDevices = async () => {
    throw new Error("network down");
  };
  vi.spyOn(console, "error").mockImplementation(() => {});

  render(<DeviceList authApi={fakeAuthApi({ listDevices })} />);

  await screen.findByText("Could not load your devices.");
});
