package realcomponents

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewProcessorRunnerAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cfg := testscommon.CreateTestConfigs(t, "../../cmd/node/config")
	pr := NewProcessorRunner(t, *cfg)
	pr.Close(t)
}
