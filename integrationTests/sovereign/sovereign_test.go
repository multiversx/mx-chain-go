package sovereign

import (
	"testing"
)

func TestChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

}
