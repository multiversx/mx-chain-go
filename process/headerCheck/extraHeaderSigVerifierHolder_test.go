package headerCheck

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/headerSigVerifier"
	"github.com/stretchr/testify/require"
)

func TestExtraHeaderSigVerifierHolder_VerifyAggregatedSignature(t *testing.T) {
	t.Parallel()

	wasVerifyCalled1 := false
	wasVerifyCalled2 := false

	expectedHdr := &block.Header{Nonce: 4}
	expectedVerifier := &cryptoMocks.MultisignerMock{}
	expectedPubKeys := [][]byte{[]byte("pk1"), []byte("pk2")}

	extraVerifier1 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		VerifyAggregatedSignatureCalled: func(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedVerifier, multiSigVerifier)
			require.Equal(t, expectedPubKeys, pubKeysSigners)

			wasVerifyCalled1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraVerifier2 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		VerifyAggregatedSignatureCalled: func(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedVerifier, multiSigVerifier)
			require.Equal(t, expectedPubKeys, pubKeysSigners)

			wasVerifyCalled2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewExtraHeaderSigVerifierHolder()
	require.False(t, holder.IsInterfaceNil())

	err := holder.RegisterExtraHeaderSigVerifier(extraVerifier1)
	require.Nil(t, err)
	err = holder.RegisterExtraHeaderSigVerifier(extraVerifier2)
	require.Nil(t, err)
	err = holder.RegisterExtraHeaderSigVerifier(extraVerifier1)
	require.Equal(t, errors.ErrExtraSignerIdAlreadyExists, err)

	err = holder.VerifyAggregatedSignature(expectedHdr, expectedVerifier, expectedPubKeys)
	require.Nil(t, err)
	require.True(t, wasVerifyCalled1)
	require.True(t, wasVerifyCalled2)
}

func TestExtraHeaderSigVerifierHolder_VerifyLeaderSignature(t *testing.T) {
	t.Parallel()

	wasVerifyCalled1 := false
	wasVerifyCalled2 := false

	expectedHdr := &block.Header{Nonce: 4}
	expectedPubKey := &mock.PublicKeyMock{}

	extraVerifier1 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		VerifyLeaderSignatureCalled: func(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedPubKey, leaderPubKey)

			wasVerifyCalled1 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraVerifier2 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		VerifyLeaderSignatureCalled: func(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error {
			require.Equal(t, expectedHdr, header)
			require.Equal(t, expectedPubKey, leaderPubKey)

			wasVerifyCalled2 = true
			return nil
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewExtraHeaderSigVerifierHolder()

	err := holder.RegisterExtraHeaderSigVerifier(extraVerifier1)
	require.Nil(t, err)
	err = holder.RegisterExtraHeaderSigVerifier(extraVerifier2)
	require.Nil(t, err)

	err = holder.VerifyLeaderSignature(expectedHdr, expectedPubKey)
	require.Nil(t, err)
	require.True(t, wasVerifyCalled1)
	require.True(t, wasVerifyCalled2)
}

func TestExtraHeaderSigVerifierHolder_RemoveLeaderSignature(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce:           4,
		Signature:       []byte("sig"),
		LeaderSignature: []byte("sig"),
	}

	extraVerifier1 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		RemoveLeaderSignatureCalled: func(header data.HeaderHandler) error {
			require.Equal(t, expectedHdr, header)
			return header.SetSignature(nil)
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraVerifier2 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		RemoveLeaderSignatureCalled: func(header data.HeaderHandler) error {
			require.Equal(t, expectedHdr, header)
			return header.SetLeaderSignature(nil)
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewExtraHeaderSigVerifierHolder()

	err := holder.RegisterExtraHeaderSigVerifier(extraVerifier1)
	require.Nil(t, err)
	err = holder.RegisterExtraHeaderSigVerifier(extraVerifier2)
	require.Nil(t, err)

	err = holder.RemoveLeaderSignature(expectedHdr)
	require.Nil(t, err)
	require.Equal(t, &block.Header{Nonce: 4}, expectedHdr)
}

func TestExtraHeaderSigVerifierHolder_RemoveAllSignatures(t *testing.T) {
	t.Parallel()

	expectedHdr := &block.Header{
		Nonce:           4,
		Signature:       []byte("sig"),
		LeaderSignature: []byte("sig"),
	}

	extraVerifier1 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		RemoveAllSignaturesCalled: func(header data.HeaderHandler) error {
			require.Equal(t, expectedHdr, header)
			return header.SetSignature(nil)
		},
		IdentifierCalled: func() string {
			return "id1"
		},
	}
	extraVerifier2 := &headerSigVerifier.ExtraHeaderSigVerifierHandlerMock{
		RemoveAllSignaturesCalled: func(header data.HeaderHandler) error {
			require.Equal(t, expectedHdr, header)
			return header.SetLeaderSignature(nil)
		},
		IdentifierCalled: func() string {
			return "id2"
		},
	}

	holder := NewExtraHeaderSigVerifierHolder()

	err := holder.RegisterExtraHeaderSigVerifier(extraVerifier1)
	require.Nil(t, err)
	err = holder.RegisterExtraHeaderSigVerifier(extraVerifier2)
	require.Nil(t, err)

	err = holder.RemoveAllSignatures(expectedHdr)
	require.Nil(t, err)
	require.Equal(t, &block.Header{Nonce: 4}, expectedHdr)
}
