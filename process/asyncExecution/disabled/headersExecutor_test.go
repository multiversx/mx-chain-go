package disabled

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeadersExecutor(t *testing.T) {
	t.Parallel()

	he := NewHeadersExecutor()
	require.NotNil(t, he)
	require.False(t, he.IsInterfaceNil())

	// Call all methods to ensure 100% coverage
	he.StartExecution()
	he.PauseExecution()
	he.ResumeExecution()

	err := he.Close()
	require.NoError(t, err)
}
