package processing

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignGenesisMetaBlockChecker(t *testing.T) {
	t.Parallel()

	checker := NewSovereignGenesisMetaBlockChecker()
	require.False(t, checker.IsInterfaceNil())
}

func TestSovereignGenesisMetaBlockChecker_CheckGenesisMetaBlock(t *testing.T) {
	t.Parallel()

	checker := NewSovereignGenesisMetaBlockChecker()
	err := checker.SetValidatorRootHashOnGenesisMetaBlock(nil, nil)
	require.Nil(t, err)

	err = checker.SetValidatorRootHashOnGenesisMetaBlock(&block.MetaBlock{}, nil)
	require.Equal(t, errors.ErrGenesisMetaBlockOnSovereign, err)
}
