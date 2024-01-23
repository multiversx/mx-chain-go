package processing

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/require"
)

func TestNewGenesisMetaBlockChecker(t *testing.T) {
	t.Parallel()

	checker := NewGenesisMetaBlockChecker()
	require.False(t, checker.IsInterfaceNil())
}

func TestGenesisMetaBlockChecker_CheckGenesisMetaBlock(t *testing.T) {
	t.Parallel()

	checker := NewGenesisMetaBlockChecker()
	hash := []byte("hash")

	err := checker.SetValidatorRootHashOnGenesisMetaBlock(nil, hash)
	require.Equal(t, errors.ErrGenesisMetaBlockDoesNotExist, err)

	err = checker.SetValidatorRootHashOnGenesisMetaBlock(&block.HeaderV2{}, hash)
	require.Equal(t, errors.ErrInvalidGenesisMetaBlock, err)

	metaBlock := &block.MetaBlock{}
	err = checker.SetValidatorRootHashOnGenesisMetaBlock(metaBlock, hash)
	require.Nil(t, err)
	require.Equal(t, &block.MetaBlock{ValidatorStatsRootHash: hash}, metaBlock)
}
