// Live updates: one WebSocket to the realtime service, which pushes chat's
// events (a new message, a new conversation, delivered and read). If the
// connection drops, it connects again by itself.
// Learn more: docs/learn/websocket.md#in-the-browser
import { fromJsonString } from "@bufbuild/protobuf";
import { ServerEventSchema } from "../gen/realtime/v1/events_pb.ts";
import type { ServerEvent } from "../gen/realtime/v1/events_pb.ts";

// What a page gets told.
export type LiveUpdateHandlers = {
  // Called for each event from the server.
  onEvent: (event: ServerEvent) => void;
  // Called each time the WebSocket opens, also after a reconnect. Events
  // sent while it was closed are lost, so the page loads what it shows
  // again here ("catching up").
  onConnected: () => void;
};

// subscribe starts the live updates, and returns a function that stops
// them (for a page that closes). Tests pass a fake subscribe.
export type Subscribe = (handlers: LiveUpdateHandlers) => () => void;

// The parts of a browser WebSocket we use. The real WebSocket has these;
// a test's fake has only these.
export type Socket = {
  onopen: ((event: Event) => void) | null;
  onmessage: ((event: MessageEvent) => void) | null;
  onclose: ((event: CloseEvent) => void) | null;
  close: () => void;
};

// After the WebSocket closes, wait this long before trying again, then
// twice as long after each failed try, up to the maximum. Waiting longer
// and longer ("exponential backoff") means that when the server is down,
// a thousand browsers don't all knock again every second.
const firstRetryDelayInMilliseconds = 1_000;
const maximumRetryDelayInMilliseconds = 30_000;

// createBrowserSocket opens a real WebSocket to the realtime service, through the
// gateway. The browser sends the session cookie with it.
function createBrowserSocket(url: string): Socket {
  return new WebSocket(url);
}

// realtimeUrl is ws://localhost:8080/realtime on our computers: the same
// address as the page, with ws: (or wss: with HTTPS) instead of http:.
function realtimeUrl(): string {
  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${scheme}//${window.location.host}/realtime`;
}

// subscribeToLiveUpdates is the real Subscribe. createSocket can be
// replaced in tests.
export function subscribeToLiveUpdates(
  handlers: LiveUpdateHandlers,
  createSocket: (url: string) => Socket = createBrowserSocket,
): () => void {
  let isStopped = false;
  let socket: Socket | null = null;
  let retryDelay = firstRetryDelayInMilliseconds;
  let retryTimer: ReturnType<typeof setTimeout> | undefined;

  function connect() {
    socket = createSocket(realtimeUrl());

    socket.onopen = () => {
      // Connected: the next drop starts waiting from the shortest delay.
      retryDelay = firstRetryDelayInMilliseconds;
      handlers.onConnected();
    };

    socket.onmessage = (message) => {
      handleMessage(message.data, handlers.onEvent);
    };

    // Called when the WebSocket closes for any reason: the server
    // restarted ("going away"), the network dropped, or it never opened.
    socket.onclose = () => {
      if (isStopped) {
        return;
      }
      retryTimer = setTimeout(connect, retryDelay);
      const doubled = retryDelay * 2;
      retryDelay = Math.min(doubled, maximumRetryDelayInMilliseconds);
    };
  }

  connect();

  return () => {
    isStopped = true;
    clearTimeout(retryTimer);
    socket?.close();
  };
}

// handleMessage reads one WebSocket message: a ServerEvent as JSON.
function handleMessage(
  data: unknown,
  onEvent: (event: ServerEvent) => void,
): void {
  if (typeof data !== "string") {
    console.error("got a websocket message that is not text");
    return;
  }

  try {
    // ignoreUnknownFields: a newer server may send fields this version of
    // the app doesn't know yet. Skip them instead of failing.
    const event = fromJsonString(ServerEventSchema, data, {
      ignoreUnknownFields: true,
    });
    onEvent(event);
  } catch (parseError) {
    console.error("could not read a websocket message", parseError);
  }
}
