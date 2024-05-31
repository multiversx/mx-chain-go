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

func TestSubRoundEndExtraSignersHolder_AggregateSignatures(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(4)
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: expectedEpoch,
		},
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			OutGoingOperationsHash: []byte("hash"),
		},
	}

	expectedBitmap := []byte("bitmap")
	expectedAggregatedSig1 := []byte("aggregatedSig1")
	expectedAggregatedSig2 := []byte("aggregatedSig2")
	extraSigner1 := &subRounds.SubRoundEndExtraSignatureMock{
		AggregateSignaturesCalled: func(bitmap []byte, header data.HeaderHandler) ([]byte, error) {
			require.Equal(t, sovHdr, header)
			require.Equal(t, expectedBitmap, bitmap)

			return expectedAggregatedSig1, nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundEndExtraSignatureMock{
		AggregateSignaturesCalled: func(bitmap []byte, header data.HeaderHandler) ([]byte, error) {
			require.Equal(t, sovHdr, header)
			require.Equal(t, expectedBitmap, bitmap)

			return expectedAggregatedSig2, nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundEndExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Equal(t, errors.ErrExtraSignerIdAlreadyExists, err)

	res, err := holder.AggregateSignatures(expectedBitmap, sovHdr)
	require.Nil(t, err)
	require.Equal(t, map[string][]byte{
		"id1": expectedAggregatedSig1,
		"id2": expectedAggregatedSig2,
	}, res)
}

func TestSubRoundEndExtraSignersHolder_AddLeaderAndAggregatedSignatures(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce: 4,
	}
	expectedCnsMsg := &consensus.Message{ChainID: []byte("1")}

	wasSigAdded1 := false
	wasSigAdded2 := false
	extraSigner1 := &subRounds.SubRoundEndExtraSignatureMock{
		AddLeaderAndAggregatedSignaturesCalled: func(header data.HeaderHandler, cnsMsg *consensus.Message) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasSigAdded1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundEndExtraSignatureMock{
		AddLeaderAndAggregatedSignaturesCalled: func(header data.HeaderHandler, cnsMsg *consensus.Message) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasSigAdded2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundEndExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.AddLeaderAndAggregatedSignatures(expectedHdr, expectedCnsMsg)
	require.Nil(t, err)
	require.True(t, wasSigAdded1)
	require.True(t, wasSigAdded2)
}

func TestSubRoundEndExtraSignersHolder_SignAndSetLeaderSignature(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce: 4,
	}
	expectedLeaderPubKey := []byte("leaderPubKey")

	wasSigAdded1 := false
	wasSigAdded2 := false
	extraSigner1 := &subRounds.SubRoundEndExtraSignatureMock{
		SignAndSetLeaderSignatureCalled: func(header data.HeaderHandler, leaderPubKey []byte) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedLeaderPubKey, leaderPubKey)

			wasSigAdded1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundEndExtraSignatureMock{
		SignAndSetLeaderSignatureCalled: func(header data.HeaderHandler, leaderPubKey []byte) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedLeaderPubKey, leaderPubKey)

			wasSigAdded2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundEndExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.SignAndSetLeaderSignature(expectedHdr, expectedLeaderPubKey)
	require.Nil(t, err)
	require.True(t, wasSigAdded1)
	require.True(t, wasSigAdded2)
}

func TestSubRoundEndExtraSignersHolder_SetAggregatedSignatureInHeader(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce: 4,
	}
	expectedAggregatedSig1 := []byte("aggregatedSig1")
	expectedAggregatedSig2 := []byte("aggregatedSig2")
	aggregatedSigs := map[string][]byte{
		"id1": expectedAggregatedSig1,
		"id2": expectedAggregatedSig2,
	}

	wasSigAdded1 := false
	wasSigAdded2 := false
	extraSigner1 := &subRounds.SubRoundEndExtraSignatureMock{
		SetAggregatedSignatureInHeaderCalled: func(header data.HeaderHandler, aggregatedSig []byte) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, aggregatedSigs["id1"], aggregatedSig)

			wasSigAdded1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundEndExtraSignatureMock{
		SetAggregatedSignatureInHeaderCalled: func(header data.HeaderHandler, aggregatedSig []byte) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, aggregatedSigs["id2"], aggregatedSig)

			wasSigAdded2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundEndExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.SetAggregatedSignatureInHeader(expectedHdr, aggregatedSigs)
	require.Nil(t, err)
	require.True(t, wasSigAdded1)
	require.True(t, wasSigAdded2)
}

func TestSubRoundEndExtraSignersHolder_VerifyAggregatedSignatures(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce: 4,
	}
	expectedBitmap := []byte("bitmap")
	wasSigVerified1 := false
	wasSigVerified2 := false
	extraSigner1 := &subRounds.SubRoundEndExtraSignatureMock{
		VerifyAggregatedSignaturesCalled: func(bitmap []byte, header data.HeaderHandler) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedBitmap, bitmap)

			wasSigVerified1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundEndExtraSignatureMock{
		VerifyAggregatedSignaturesCalled: func(bitmap []byte, header data.HeaderHandler) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedBitmap, bitmap)

			wasSigVerified2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundEndExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.VerifyAggregatedSignatures(expectedHdr, expectedBitmap)
	require.Nil(t, err)
	require.True(t, wasSigVerified1)
	require.True(t, wasSigVerified2)
}

func TestSubRoundEndExtraSignersHolder_HaveConsensusHeaderWithFullInfo(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce: 4,
	}
	expectedCnsMsg := &consensus.Message{ChainID: []byte("1")}

	wasInfoAdded1 := false
	wasInfoAdded2 := false
	extraSigner1 := &subRounds.SubRoundEndExtraSignatureMock{
		HaveConsensusHeaderWithFullInfoCalled: func(header data.HeaderHandler, cnsMsg *consensus.Message) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasInfoAdded1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraSigner2 := &subRounds.SubRoundEndExtraSignatureMock{
		HaveConsensusHeaderWithFullInfoCalled: func(header data.HeaderHandler, cnsMsg *consensus.Message) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedCnsMsg, cnsMsg)

			wasInfoAdded2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewSubRoundEndExtraSignersHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraSigningHandler(extraSigner1)
	require.Nil(t, err)
	err = holder.RegisterExtraSigningHandler(extraSigner2)
	require.Nil(t, err)

	err = holder.HaveConsensusHeaderWithFullInfo(expectedHdr, expectedCnsMsg)
	require.Nil(t, err)
	require.True(t, wasInfoAdded1)
	require.True(t, wasInfoAdded2)
}
