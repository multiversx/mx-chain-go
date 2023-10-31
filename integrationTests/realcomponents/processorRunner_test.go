package realcomponents

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewProcessorRunnerAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cfg, err := testscommon.CreateTestConfigs("../../cmd/node/config")
	require.Nil(t, err)

	pr := NewProcessorRunner(t, *cfg)
	pr.Close(t)
}
