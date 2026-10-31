// @vitest-environment jsdom
// Tests for the whole app: which page you see, logging in, signing up and
// logging out. Learn more: docs/learn/testing.md and docs/learn/react-router.md
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ConnectError, Code } from "@connectrpc/connect";
import { MemoryRouter } from "react-router";
import { afterEach, expect, test, vi } from "vitest";
import App from "./App.tsx";
import type { AuthApi, User } from "./auth/authClient.ts";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

const sadra: User = {
  id: "019a3f2c-7b1e-7c4d-9a2b-3c4d5e6f7a8b",
  username: "sadra",
};

// fakeAuthApi returns an AuthApi where nobody is logged in and every call
// works. A test replaces the calls it cares about.
function fakeAuthApi(changes: Partial<AuthApi> = {}): AuthApi {
  const defaults: AuthApi = {
    signUp: async () => {},
    logIn: async () => sadra,
    getCurrentUser: async () => null,
    logOut: async () => {},
    listDevices: async () => [],
    logOutDevice: async () => {},
  };
  // "...": copy all fields of defaults, then overwrite them with changes.
  return { ...defaults, ...changes };
}

const loadGreeting = async () => ({ id: 1n, message: "Hello, test!" });

// renderApp draws the app as if the browser had opened this address.
// MemoryRouter keeps the address in memory instead of the address bar.
function renderApp(authApi: AuthApi, address: string) {
  render(
    <MemoryRouter initialEntries={[address]}>
      <App authApi={authApi} loadGreeting={loadGreeting} />
    </MemoryRouter>,
  );
}

// typeInto types a value into the input with this label.
function typeInto(label: string, value: string) {
  const input = screen.getByLabelText(label);
  fireEvent.change(input, { target: { value } });
}

test("sends you to the login page when nobody is logged in", async () => {
  renderApp(fakeAuthApi(), "/account");

  await screen.findByRole("heading", { name: "Log in" });
});

test("sends you to your account when you are logged in", async () => {
  const authApi = fakeAuthApi({ getCurrentUser: async () => sadra });

  renderApp(authApi, "/login");

  await screen.findByRole("heading", { name: "Your account" });
  screen.getByText("sadra");
});

test("logs you in and shows your account", async () => {
  const logIn = vi.fn(async () => sadra);
  renderApp(fakeAuthApi({ logIn }), "/login");
  await screen.findByRole("heading", { name: "Log in" });

  typeInto("Username", "sadra");
  typeInto("Password", "purple tram winter");
  fireEvent.click(screen.getByRole("button", { name: "Log in" }));

  await screen.findByRole("heading", { name: "Your account" });
  expect(logIn).toHaveBeenCalledWith("sadra", "purple tram winter");
});

test("shows the service's message when the login fails", async () => {
  const logIn = async () => {
    throw new ConnectError("wrong username or password", Code.Unauthenticated);
  };
  renderApp(fakeAuthApi({ logIn }), "/login");
  await screen.findByRole("heading", { name: "Log in" });

  typeInto("Username", "sadra");
  typeInto("Password", "purple tram summer");
  fireEvent.click(screen.getByRole("button", { name: "Log in" }));

  await screen.findByText("Wrong username or password");
});

test("hides the details of unexpected errors", async () => {
  const logIn = async () => {
    throw new ConnectError("database connection refused", Code.Internal);
  };
  renderApp(fakeAuthApi({ logIn }), "/login");
  await screen.findByRole("heading", { name: "Log in" });

  typeInto("Username", "sadra");
  typeInto("Password", "purple tram winter");
  fireEvent.click(screen.getByRole("button", { name: "Log in" }));

  await screen.findByText("Something went wrong. Please try again.");
  expect(screen.queryByText(/database/)).toBeNull();
});

test("signs you up, logs you in and shows your account", async () => {
  const signUp = vi.fn(async () => {});
  const logIn = vi.fn(async () => sadra);
  renderApp(fakeAuthApi({ signUp, logIn }), "/signup");
  await screen.findByRole("heading", { name: "Create an account" });

  typeInto("Username", "sadra");
  typeInto("Password", "purple tram winter");
  fireEvent.click(screen.getByRole("button", { name: "Create account" }));

  await screen.findByRole("heading", { name: "Your account" });
  expect(signUp).toHaveBeenCalledWith("sadra", "purple tram winter");
  expect(logIn).toHaveBeenCalledWith("sadra", "purple tram winter");
});

test("logs you out and shows the login page", async () => {
  const logOut = vi.fn(async () => {});
  const authApi = fakeAuthApi({ getCurrentUser: async () => sadra, logOut });
  renderApp(authApi, "/account");
  await screen.findByRole("heading", { name: "Your account" });

  fireEvent.click(screen.getByRole("button", { name: "Log out" }));

  await screen.findByRole("heading", { name: "Log in" });
  expect(logOut).toHaveBeenCalled();
});

test("still shows the greeting", async () => {
  renderApp(fakeAuthApi(), "/login");

  await screen.findByText("Hello, test!");
});
