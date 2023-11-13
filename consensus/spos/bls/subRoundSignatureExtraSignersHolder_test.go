package bls

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
	"github.com/stretchr/testify/require"
)

func TestSubRoundSignatureExtraSignersHolder_CreateExtraSignatureShares(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.SovereignChainHeader{}
	expectedSelfIndex := uint16(4)
	expectedSelfPubKey := []byte("selfPubKey")
	extraSigner1 := &subRounds.SubRoundSignatureExtraSignatureHandlerMock{
		CreateSignatureShareCalled: func(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error) {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedSelfIndex, selfIndex)
			require.Equal(t, expectedSelfPubKey, selfPubKey)

			return []byte("sigShare1"), nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundSignatureExtraSignatureHandlerMock{
		CreateSignatureShareCalled: func(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error) {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedSelfIndex, selfIndex)
			require.Equal(t, expectedSelfPubKey, selfPubKey)

			return []byte("sigShare2"), nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundSignatureExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Equal(t, errors.ErrExtraSignerIdAlreadyExists, err)

	res, err := holder.CreateExtraSignatureShares(expectedHdr, expectedSelfIndex, expectedSelfPubKey)
	require.Nil(t, err)
	require.Equal(t, map[string][]byte{
		"id1": []byte("sigShare1"),
		"id2": []byte("sigShare2"),
	}, res)
}

func TestSubRoundSignatureExtraSignersHolder_AddExtraSigSharesToConsensusMessage(t *testing.T) {
	t.Parallel()

	expectedCnsMsg := &consensus.Message{ChainID: []byte("1")}
	expectedSigShares := map[string][]byte{
		"id1": []byte("sigShare1"),
		"id2": []byte("sigShare2"),
	}

	wasAdded1 := false
	wasAdded2 := false
	extraSigner1 := &subRounds.SubRoundSignatureExtraSignatureHandlerMock{
		AddSigShareToConsensusMessageCalled: func(sigShare []byte, cnsMsg *consensus.Message) {
			require.Equal(t, []byte("sigShare1"), sigShare)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasAdded1 = true
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundSignatureExtraSignatureHandlerMock{
		AddSigShareToConsensusMessageCalled: func(sigShare []byte, cnsMsg *consensus.Message) {
			require.Equal(t, []byte("sigShare2"), sigShare)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasAdded2 = true
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundSignatureExtraSignersHolder()
	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.AddExtraSigSharesToConsensusMessage(expectedSigShares, expectedCnsMsg)
	require.Nil(t, err)
	require.True(t, wasAdded1)
	require.True(t, wasAdded2)
}

func TestSubRoundSignatureExtraSignersHolder_StoreExtraSignatureShare(t *testing.T) {
	t.Parallel()

	expectedSelfIndex := uint16(4)
	expectedCnsMsg := &consensus.Message{ChainID: []byte("1")}

	wasStored1 := false
	wasStored2 := false
	extraSigner1 := &subRounds.SubRoundSignatureExtraSignatureHandlerMock{
		StoreSignatureShareCalled: func(index uint16, cnsMsg *consensus.Message) error {
			require.Equal(t, expectedSelfIndex, index)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasStored1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundSignatureExtraSignatureHandlerMock{
		StoreSignatureShareCalled: func(index uint16, cnsMsg *consensus.Message) error {
			require.Equal(t, expectedSelfIndex, index)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasStored2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundSignatureExtraSignersHolder()
	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.StoreExtraSignatureShare(expectedSelfIndex, expectedCnsMsg)
	require.Nil(t, err)
	require.True(t, wasStored1)
	require.True(t, wasStored2)
}
