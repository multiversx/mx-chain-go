package p2p

import "testing"

func TestFailNewP2PParams(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	//invalid port
	NewP2PParams(65536)
}
