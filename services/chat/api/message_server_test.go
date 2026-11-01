package api

import (
	"testing"

	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
)

func TestToPositionReadsTheOneof(t *testing.T) {
	testCases := []struct {
		name         string
		request      *chatv1.ListMessagesRequest
		wantPosition conversation.Position
		wantSequence int64
	}{
		{"none set", &chatv1.ListMessagesRequest{}, conversation.PositionNewest, 0},
		{
			"after",
			&chatv1.ListMessagesRequest{Position: &chatv1.ListMessagesRequest_AfterSequence{AfterSequence: 41}},
			conversation.PositionAfter, 41,
		},
		{
			// 0 is a real choice here: "after 0" means from the first message,
			// not "nothing set". A oneof can tell the two apart.
			"after zero",
			&chatv1.ListMessagesRequest{Position: &chatv1.ListMessagesRequest_AfterSequence{AfterSequence: 0}},
			conversation.PositionAfter, 0,
		},
		{
			"before",
			&chatv1.ListMessagesRequest{Position: &chatv1.ListMessagesRequest_BeforeSequence{BeforeSequence: 80}},
			conversation.PositionBefore, 80,
		},
	}

	for _, testCase := range testCases {
		position, sequence := toPosition(testCase.request)

		if position != testCase.wantPosition || sequence != testCase.wantSequence {
			t.Errorf("%s: got (%v, %d), want (%v, %d)", testCase.name, position, sequence, testCase.wantPosition, testCase.wantSequence)
		}
	}
}
