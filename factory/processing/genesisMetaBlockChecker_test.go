package processing

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

	genesisBlocks := map[uint32]data.HeaderHandler{
		0: &block.HeaderV2{},
	}
	err := checker.CheckGenesisMetaBlock(genesisBlocks, []byte("hash"))
	require.Equal(t, errGenesisMetaBlockDoesNotExist, err)

	genesisBlocks = map[uint32]data.HeaderHandler{
		core.MetachainShardId: &block.HeaderV2{},
	}
	err = checker.CheckGenesisMetaBlock(genesisBlocks, []byte("hash"))
	require.Equal(t, errInvalidGenesisMetaBlock, err)

	genesisBlocks = map[uint32]data.HeaderHandler{
		core.MetachainShardId: &block.MetaBlock{},
	}
	err = checker.CheckGenesisMetaBlock(genesisBlocks, []byte("hash"))
	require.Nil(t, err)
}
