package realcomponents

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewProcessorRunnerAndClose(t *testing.T) {
	cfg := testscommon.CreateTestConfigs(t, "../../cmd/node/config")
	pr := NewProcessorRunner(t, *cfg)
	pr.Close(t)
}
