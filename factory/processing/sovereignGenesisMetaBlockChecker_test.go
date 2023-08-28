package processing

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/errors"
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
	err := checker.SetValidatorRootHashOnGenesisMetaBlock(nil, nil)
	require.Nil(t, err)

	err = checker.SetValidatorRootHashOnGenesisMetaBlock(&block.MetaBlock{}, nil)
	require.Equal(t, errors.ErrGenesisMetaBlockOnSovereign, err)
}
