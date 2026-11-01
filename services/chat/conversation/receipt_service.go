package conversation

// Delivered and read: the two watermarks of each member (ADR 0032).
// Learn more: docs/learn/message-ordering.md#delivered-and-read

import (
	"context"
	"fmt"
)

// MarkDelivered records that the caller's device received the
// conversation's messages up to this sequence number.
func (service *Service) MarkDelivered(ctx context.Context, callerID string, conversationID string, sequence int64) error {
	err := service.checkReceipt(ctx, callerID, conversationID, sequence)
	if err != nil {
		return err
	}

	err = service.store.MarkDelivered(ctx, conversationID, callerID, sequence)
	if err != nil {
		return fmt.Errorf("mark delivered: %w", err)
	}
	return nil
}

// MarkRead records that the caller has seen the conversation's messages up
// to this sequence number.
func (service *Service) MarkRead(ctx context.Context, callerID string, conversationID string, sequence int64) error {
	err := service.checkReceipt(ctx, callerID, conversationID, sequence)
	if err != nil {
		return err
	}

	err = service.store.MarkRead(ctx, conversationID, callerID, sequence)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

// checkReceipt checks the rules both calls share: a real sequence number,
// and a conversation the caller is in.
func (service *Service) checkReceipt(ctx context.Context, callerID string, conversationID string, sequence int64) error {
	// Messages start at 1, so 0 would mark nothing.
	if sequence < 1 {
		return invalidRequest("the sequence number must be 1 or more")
	}
	return service.checkMember(ctx, callerID, conversationID)
}
