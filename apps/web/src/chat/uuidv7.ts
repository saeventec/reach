// Makes UUIDv7 ids for new messages. The app makes each message's id
// before sending it, so a retry sends the same id, and the server stores
// the message only once (ADR 0032).
// Learn more: docs/learn/message-ordering.md#idempotency-ids-made-by-the-app
//
// A UUIDv7 is 16 bytes: the first 6 are the time in milliseconds, so newer
// ids sort after older ones; the rest is random, apart from two small
// markers: the version (7) and the variant (binary 10). The browser has
// crypto.randomUUID(), but it makes version 4 (all random), so we make v7
// ourselves. The layout is in RFC 9562.

// newUUIDv7 returns an id like "01a19641-0c1e-7a2b-9c3d-4e5f6a7b8c9d".
// The two values can be set by tests; normally they are the current time
// and 10 random bytes from the browser's secure random generator.
export function newUUIDv7(
  nowInMilliseconds: number = Date.now(),
  randomBytes: Uint8Array = crypto.getRandomValues(new Uint8Array(10)),
): string {
  const bytes = new Uint8Array(16);

  // Bytes 0 to 5: the time, the most significant byte first. A
  // millisecond time needs 48 bits; JavaScript numbers hold 53 exactly.
  let time = nowInMilliseconds;
  for (let index = 5; index >= 0; index--) {
    bytes[index] = time % 256;
    time = Math.floor(time / 256);
  }

  // Bytes 6 to 15: random.
  bytes.set(randomBytes, 6);

  // The version: the top 4 bits of byte 6 become 0111 (7).
  bytes[6] = (bytes[6] & 0x0f) | 0x70;
  // The variant: the top 2 bits of byte 8 become 10.
  bytes[8] = (bytes[8] & 0x3f) | 0x80;

  return formatUUID(bytes);
}

// formatUUID writes 16 bytes as 32 hexadecimal digits in the groups
// 8-4-4-4-12.
function formatUUID(bytes: Uint8Array): string {
  const digits = Array.from(bytes, (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
  const groups = [
    digits.slice(0, 8),
    digits.slice(8, 12),
    digits.slice(12, 16),
    digits.slice(16, 20),
    digits.slice(20, 32),
  ];
  return groups.join("-");
}
