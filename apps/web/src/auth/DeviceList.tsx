// The list of devices the user is logged in on, with a button to log out
// each of the other ones (for example, a lost phone).
// Learn more: docs/learn/sessions.md#devices
import { useEffect, useState } from "react";
import { errorMessage } from "../shared/errorMessage.ts";
import type { AuthApi, Device } from "./authClient.ts";

type DevicesState =
  | { status: "loading" }
  | { status: "loaded"; devices: Device[] }
  | { status: "failed" };

type DeviceListProps = {
  authApi: AuthApi;
};

export function DeviceList({ authApi }: DeviceListProps) {
  const [state, setState] = useState<DevicesState>({ status: "loading" });
  const [error, setError] = useState("");

  useEffect(() => {
    // See NewestGreeting.tsx for why "ignore" is needed.
    let ignore = false;

    async function loadDevices() {
      try {
        const devices = await authApi.listDevices();
        if (!ignore) {
          setState({ status: "loaded", devices });
        }
      } catch (loadError) {
        console.error("could not load the devices", loadError);
        if (!ignore) {
          setState({ status: "failed" });
        }
      }
    }

    loadDevices();

    return () => {
      ignore = true;
    };
  }, [authApi]);

  async function handleLogOutDevice(deviceId: string) {
    setError("");
    try {
      await authApi.logOutDevice(deviceId);
      // Remove the device from the list on the screen, without loading the
      // whole list again.
      setState((current) => removeDevice(current, deviceId));
    } catch (logOutError) {
      setError(errorMessage(logOutError));
    }
  }

  if (state.status === "loading") {
    return <p className="text-slate-500">Loading your devices…</p>;
  }

  if (state.status === "failed") {
    return <p className="text-red-600">Could not load your devices.</p>;
  }

  return (
    <div className="flex flex-col gap-2">
      <h3 className="text-lg font-semibold text-slate-900">Your devices</h3>
      {error !== "" && <p className="text-red-600">{error}</p>}
      <ul className="flex flex-col gap-2">
        {state.devices.map((device) => (
          // "key" helps React tell the items apart when the list changes.
          <DeviceItem
            key={device.id}
            device={device}
            onLogOut={() => handleLogOutDevice(device.id)}
          />
        ))}
      </ul>
    </div>
  );
}

// removeDevice returns the state without one device.
function removeDevice(state: DevicesState, deviceId: string): DevicesState {
  if (state.status !== "loaded") {
    return state;
  }
  const remaining = state.devices.filter((device) => device.id !== deviceId);
  return { status: "loaded", devices: remaining };
}

type DeviceItemProps = {
  device: Device;
  onLogOut: () => void;
};

function DeviceItem({ device, onLogOut }: DeviceItemProps) {
  const lastUsed = formatDate(device.lastUsedAt);

  return (
    <li className="flex items-center justify-between rounded border border-slate-200 bg-white px-3 py-2">
      <div>
        <p className="text-slate-900">{device.name}</p>
        <p className="text-xs text-slate-500">Last used {lastUsed}</p>
      </div>
      {device.isCurrentDevice ? (
        <span className="text-xs text-blue-600">This device</span>
      ) : (
        <button
          type="button"
          onClick={onLogOut}
          aria-label={`Log out ${device.name}`}
          className="text-sm text-red-600 hover:underline"
        >
          Log out
        </button>
      )}
    </li>
  );
}

// formatDate writes a date the way the person's browser language does,
// like "Oct 31, 2026, 9:00 AM".
function formatDate(date: Date): string {
  const format = new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
  return format.format(date);
}
