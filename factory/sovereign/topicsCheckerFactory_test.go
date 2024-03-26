package sovereign_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/factory/sovereign"
)

func TestNewTopicsCheckerFactory(t *testing.T) {
	t.Parallel()

	topicsCheckerFactory := sovereign.NewTopicsCheckerFactory()
	require.False(t, topicsCheckerFactory.IsInterfaceNil())
}

func TestTopicsCheckerFactory_CreateTopicsChecker(t *testing.T) {
	t.Parallel()

	topicsCheckerFactory := sovereign.NewTopicsCheckerFactory()
	topicsChecker := topicsCheckerFactory.CreateTopicsChecker()
	require.NotNil(t, topicsChecker)
	require.Equal(t, "*disabled.topicsChecker", fmt.Sprintf("%T", topicsChecker))
}
