package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateProcessDebugger(t *testing.T) {
	t.Parallel()

	t.Run("create disabled process debugger", func(t *testing.T) {
		t.Parallel()

		configs := config.ProcessDebugConfig{
			Enabled: false,
		}
		debugger, err := CreateProcessDebugger(configs)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(debugger))
		assert.Equal(t, "*process.disabledDebugger", fmt.Sprintf("%T", debugger))
	})
	t.Run("create real process debugger", func(t *testing.T) {
		t.Parallel()

		configs := config.ProcessDebugConfig{
			Enabled:              true,
			DebuggingLogLevel:    "",
			GoRoutinesDump:       false,
			PollingTimeInSeconds: 1,
		}
		debugger, err := CreateProcessDebugger(configs)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(debugger))
		assert.Equal(t, "*process.processDebugger", fmt.Sprintf("%T", debugger))
	})
	t.Run("create real process debugger errors", func(t *testing.T) {
		t.Parallel()

		configs := config.ProcessDebugConfig{
			Enabled: true,
		}
		debugger, err := CreateProcessDebugger(configs)
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(debugger))
	})
}
