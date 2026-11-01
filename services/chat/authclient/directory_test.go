package authclient

import (
	"fmt"
	"testing"
)

func TestSplitIntoParts(t *testing.T) {
	testCases := []struct {
		count     int
		wantSizes []int
	}{
		{0, []int{}},
		{1, []int{1}},
		{100, []int{100}},
		{101, []int{100, 1}},
		{250, []int{100, 100, 50}},
	}

	for _, testCase := range testCases {
		values := make([]string, 0, testCase.count)
		for number := range testCase.count {
			values = append(values, fmt.Sprintf("value%d", number))
		}

		parts := splitIntoParts(values)

		sizes := []int{}
		for _, part := range parts {
			sizes = append(sizes, len(part))
		}
		if fmt.Sprint(sizes) != fmt.Sprint(testCase.wantSizes) {
			t.Errorf("%d values: got parts of %v, want %v", testCase.count, sizes, testCase.wantSizes)
		}
	}
}
