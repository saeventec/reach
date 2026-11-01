import { expect, test } from "vitest";
import { newUUIDv7 } from "./uuidv7.ts";

// The same pattern the chat service checks (message_rules.go): version 7,
// and the variant digit 8, 9, a or b.
const uuidVersion7Pattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

test("makes ids the chat service accepts", () => {
  for (let count = 0; count < 100; count++) {
    expect(newUUIDv7()).toMatch(uuidVersion7Pattern);
  }
});

test("starts with the time, so newer ids sort after older ones", () => {
  const noRandom = new Uint8Array(10);
  const earlier = newUUIDv7(Date.UTC(2026, 10, 1, 12, 0, 0), noRandom);
  const later = newUUIDv7(Date.UTC(2026, 10, 1, 12, 0, 1), noRandom);

  // 2026-11-01 12:00:00 UTC is 0x01a196fb1600 milliseconds since 1970:
  // the first 12 digits. With no random bits, the rest is only the markers.
  expect(earlier).toBe("01a196fb-1600-7000-8000-000000000000");
  // One second later is 1000 (0x3e8) more: 0x01a196fb19e8.
  expect(later).toBe("01a196fb-19e8-7000-8000-000000000000");
  expect(later > earlier).toBe(true);
});

test("sets the version and variant even over random bytes that are all 1", () => {
  const allOnes = new Uint8Array(10).fill(0xff);

  const id = newUUIDv7(0, allOnes);

  expect(id).toBe("00000000-0000-7fff-bfff-ffffffffffff");
});

test("two ids made at the same moment are different", () => {
  const now = Date.now();

  expect(newUUIDv7(now)).not.toBe(newUUIDv7(now));
});
