package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	blproc "github.com/multiversx/mx-chain-go/process/block"
)

func TestMetaProcessor_VerifyBlockProposal(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	mp, err := blproc.NewMetaProcessor(arguments)
	require.Nil(t, err)

	header := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
	}
	body := &block.Body{}

	err = mp.VerifyBlockProposal(header, body, haveTime)
	require.NoError(t, err)
}

func TestMetaProcessor_OnProposedBlock(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	mp, err := blproc.NewMetaProcessor(arguments)
	require.Nil(t, err)

	err = mp.OnProposedBlock(nil, nil, nil)
	require.NoError(t, err)
}
