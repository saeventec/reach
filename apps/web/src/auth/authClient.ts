// Talks to the auth service (AccountService and SessionService), with code
// generated from proto/auth/v1/. Learn more: docs/learn/connect-rpc.md
//
// The session token travels in an HttpOnly cookie that this code can't
// read. The browser adds it to every call by itself, because the calls go
// to the same address as the page. Learn more: docs/learn/sessions.md
import { timestampDate } from "@bufbuild/protobuf/wkt";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AccountService } from "../gen/auth/v1/account_pb.ts";
import { SessionService } from "../gen/auth/v1/session_pb.ts";
import type { Device as DeviceMessage } from "../gen/auth/v1/session_pb.ts";

// A user, as the rest of the web app sees it.
export type User = {
  id: string;
  username: string;
};

// A device the user is logged in on: one session.
export type Device = {
  id: string;
  // A readable guess, like "Chrome on Windows".
  name: string;
  loggedInAt: Date;
  lastUsedAt: Date;
  // True for the device showing the page.
  isCurrentDevice: boolean;
};

// All the calls the pages need. The app passes the real ones (authApi
// below); tests pass fakes. Learn more: docs/learn/testing.md#fakes
export type AuthApi = {
  signUp: (username: string, password: string) => Promise<void>;
  logIn: (username: string, password: string) => Promise<User>;
  // Returns null when nobody is logged in.
  getCurrentUser: () => Promise<User | null>;
  logOut: () => Promise<void>;
  listDevices: () => Promise<Device[]>;
  logOutDevice: (deviceId: string) => Promise<void>;
};

const transport = createConnectTransport({ baseUrl: window.location.origin });
const accountClient = createClient(AccountService, transport);
const sessionClient = createClient(SessionService, transport);

async function signUp(username: string, password: string): Promise<void> {
  await accountClient.signUp({ username, password });
}

async function logIn(username: string, password: string): Promise<User> {
  const response = await sessionClient.logIn({ username, password });
  return { id: response.userId, username: response.username };
}

async function getCurrentUser(): Promise<User | null> {
  try {
    const response = await sessionClient.getCurrentUser({});
    return { id: response.userId, username: response.username };
  } catch (error) {
    // "Not logged in" is a normal answer here, not a failure.
    const connectError = ConnectError.from(error);
    if (connectError.code === Code.Unauthenticated) {
      return null;
    }
    throw error;
  }
}

async function logOut(): Promise<void> {
  await sessionClient.logOut({});
}

async function listDevices(): Promise<Device[]> {
  const response = await sessionClient.listDevices({});
  const devices = response.devices.map(toDevice);
  return devices;
}

async function logOutDevice(deviceId: string): Promise<void> {
  await sessionClient.logOutDevice({ deviceId });
}

// toDevice turns a Device message from the API into the app's Device type,
// with plain JavaScript dates.
function toDevice(message: DeviceMessage): Device {
  return {
    id: message.id,
    name: message.name,
    loggedInAt: toDate(message.loggedInAt),
    lastUsedAt: toDate(message.lastUsedAt),
    isCurrentDevice: message.isCurrentDevice,
  };
}

// toDate turns a Protobuf Timestamp into a JavaScript Date. In Protobuf, a
// message field can always be missing, so TypeScript makes us handle that.
// The service always sends the dates, so a missing one is a bug.
function toDate(timestamp: Timestamp | undefined): Date {
  if (timestamp === undefined) {
    throw new Error("the auth service sent a device without a date");
  }
  return timestampDate(timestamp);
}

// The real API calls, for the app.
export const authApi: AuthApi = {
  signUp,
  logIn,
  getCurrentUser,
  logOut,
  listDevices,
  logOutDevice,
};
