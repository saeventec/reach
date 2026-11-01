// Delivered and read, from the members' watermarks (ADR 0032).
// Learn more: docs/learn/message-ordering.md#delivered-and-read
import type { Member, Message } from "./chatClient.ts";

// The status of a message you sent: "sent" (✓), "delivered" to everyone
// else (✓✓), or "read" by everyone else.
export type MessageStatus = "sent" | "delivered" | "read";

// messageStatus compares the message's sequence number with the other
// members' watermarks. In a group, it needs all of them: "delivered"
// means delivered to everyone.
export function messageStatus(
  message: Message,
  members: Member[],
): MessageStatus {
  const others = members.filter((member) => member.userId !== message.senderId);
  if (others.length === 0) {
    return "sent";
  }

  const isRead = others.every((member) => member.readUpTo >= message.sequence);
  if (isRead) {
    return "read";
  }
  const isDelivered = others.every(
    (member) => member.deliveredUpTo >= message.sequence,
  );
  if (isDelivered) {
    return "delivered";
  }
  return "sent";
}

// A member's new watermarks, from a live receipt event.
export type Receipt = {
  userId: string;
  deliveredUpTo: bigint;
  readUpTo: bigint;
};

// applyReceipt returns the members with one member's new watermarks.
// They only go up, like on the server, so a late event can't move them
// back.
export function applyReceipt(members: Member[], receipt: Receipt): Member[] {
  return members.map((member) => {
    if (member.userId !== receipt.userId) {
      return member;
    }
    return {
      ...member,
      deliveredUpTo: larger(member.deliveredUpTo, receipt.deliveredUpTo),
      readUpTo: larger(member.readUpTo, receipt.readUpTo),
    };
  });
}

// mergeMembers takes freshly loaded members (after a reconnect, when
// receipt events may have been missed) and keeps, for each one, the higher
// of the two watermarks: a fresh load can be a moment older than a live
// event that arrived meanwhile, and watermarks never go down.
export function mergeMembers(current: Member[], fresh: Member[]): Member[] {
  return fresh.map((freshMember) => {
    const known = current.find(
      (member) => member.userId === freshMember.userId,
    );
    if (known === undefined) {
      return freshMember;
    }
    return {
      ...freshMember,
      deliveredUpTo: larger(known.deliveredUpTo, freshMember.deliveredUpTo),
      readUpTo: larger(known.readUpTo, freshMember.readUpTo),
    };
  });
}

// larger returns the larger of two bigints. (Math.max only takes numbers.)
function larger(first: bigint, second: bigint): bigint {
  return first > second ? first : second;
}
