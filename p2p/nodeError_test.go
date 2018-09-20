package p2p

import "testing"

func TestNodeError(t *testing.T) {
	err := NewNodeError("A", "B", "MES")

	if err.Err != err.Error() {
		t.Fatal("Node error should return Err on Error()")
	}
}
