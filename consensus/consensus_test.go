package consensus

import (
	"testing"
)

func TestResponseType(t *testing.T) {
	if RT_AGREE != 0 || RT_DISAGREE != 1 || RT_DISAGREE != 2 || RT_NOT_AVAILABLE != 3 {
		t.Fatal("Wrong values in answer type enum")
	}
}
