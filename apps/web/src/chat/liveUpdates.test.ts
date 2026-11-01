// @vitest-environment jsdom
import { afterEach, expect, test, vi } from "vitest";
import type { ServerEvent } from "../gen/realtime/v1/events_pb.ts";
import { subscribeToLiveUpdates } from "./liveUpdates.ts";
import type { Socket } from "./liveUpdates.ts";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

// FakeSocket stands in for a WebSocket. The test calls open, receive and
// drop, as the server or the network would.
class FakeSocket implements Socket {
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  isClosed = false;
  url: string;

  constructor(url: string) {
    this.url = url;
  }

  close() {
    this.isClosed = true;
  }

  open() {
    this.onopen?.(new Event("open"));
  }

  receive(data: string) {
    this.onmessage?.(new MessageEvent("message", { data }));
  }

  drop() {
    this.onclose?.(new CloseEvent("close"));
  }
}

// start subscribes with fake sockets, and returns what happened.
function start() {
  const sockets: FakeSocket[] = [];
  const events: ServerEvent[] = [];
  const onConnected = vi.fn();
  const createSocket = (url: string) => {
    const socket = new FakeSocket(url);
    sockets.push(socket);
    return socket;
  };
  const handlers = {
    onEvent: (event: ServerEvent) => events.push(event),
    onConnected,
  };
  const stop = subscribeToLiveUpdates(handlers, createSocket);
  return { sockets, events, onConnected, stop };
}

test("connects to /realtime on the page's own address", () => {
  const { sockets } = start();

  expect(sockets[0].url).toBe(`ws://${window.location.host}/realtime`);
});

test("passes on each event, read from JSON", () => {
  const { sockets, events } = start();
  sockets[0].open();

  sockets[0].receive(
    '{"conversationCreated": {"conversationId": "chat-1", "memberIds": ["alice"]}}',
  );

  expect(events).toHaveLength(1);
  expect(events[0].event.case).toBe("conversationCreated");
});

test("skips a message it can't read", () => {
  const { sockets, events } = start();
  vi.spyOn(console, "error").mockImplementation(() => {});
  sockets[0].open();

  sockets[0].receive("not json");

  expect(events).toHaveLength(0);
});

test("tells the page on every connect, so it can catch up", () => {
  vi.useFakeTimers();
  const { sockets, onConnected } = start();
  sockets[0].open();

  sockets[0].drop();
  vi.advanceTimersByTime(1_000);
  sockets[1].open();

  expect(onConnected).toHaveBeenCalledTimes(2);
});

test("waits longer after each failed try, then starts over", () => {
  // Fake timers: the test moves the clock itself, instead of waiting.
  vi.useFakeTimers();
  const { sockets } = start();

  sockets[0].drop();
  vi.advanceTimersByTime(999);
  expect(sockets).toHaveLength(1); // not yet: the first wait is 1 second
  vi.advanceTimersByTime(1);
  expect(sockets).toHaveLength(2);

  sockets[1].drop();
  vi.advanceTimersByTime(1_999);
  expect(sockets).toHaveLength(2); // the second wait is 2 seconds
  vi.advanceTimersByTime(1);
  expect(sockets).toHaveLength(3);

  // Once a try works, the next drop waits 1 second again.
  sockets[2].open();
  sockets[2].drop();
  vi.advanceTimersByTime(1_000);
  expect(sockets).toHaveLength(4);
});

test("stops for good when the page closes", () => {
  vi.useFakeTimers();
  const { sockets, stop } = start();
  sockets[0].open();

  stop();
  sockets[0].drop();
  vi.advanceTimersByTime(60_000);

  expect(sockets[0].isClosed).toBe(true);
  expect(sockets).toHaveLength(1);
});
