package processing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDisabledGenesisMetaBlockChecker(t *testing.T) {
	t.Parallel()

	checker := NewDisabledGenesisMetaBlockChecker()
	require.False(t, checker.IsInterfaceNil())
}

func TestDisabledGenesisMetaBlockChecker_CheckGenesisMetaBlock(t *testing.T) {
	t.Parallel()

	checker := NewDisabledGenesisMetaBlockChecker()
	require.Nil(t, checker.CheckGenesisMetaBlock(nil, nil))
}
