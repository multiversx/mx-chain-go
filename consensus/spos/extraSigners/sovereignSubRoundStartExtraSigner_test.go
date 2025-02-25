package extraSigners

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/spos"
	cnsTest "github.com/multiversx/mx-chain-go/testscommon/consensus"
)

func TestNewSovereignSubRoundStartOutGoingTxData(t *testing.T) {
	t.Parallel()

	t.Run("nil signing handler, should return error", func(t *testing.T) {
		sovSigHandler, err := NewSovereignSubRoundStartExtraSigner(nil, block.OutGoingMbTx)
		require.Equal(t, spos.ErrNilSigningHandler, err)
		require.True(t, check.IfNil(sovSigHandler))
	})

	t.Run("should work", func(t *testing.T) {
		sovSigHandler, err := NewSovereignSubRoundStartExtraSigner(&cnsTest.SigningHandlerStub{}, block.OutGoingMbTx)
		require.Nil(t, err)
		require.False(t, sovSigHandler.IsInterfaceNil())
	})
}
func TestSovereignSubRoundStartOutGoingTxData_Reset(t *testing.T) {
	t.Parallel()

	expectedPubKeys := []string{"pk1", "pk2"}
	wasResetCalled := false
	sigHandler := &cnsTest.SigningHandlerStub{
		ResetCalled: func(pubKeys []string) error {
			require.Equal(t, expectedPubKeys, pubKeys)

			wasResetCalled = true
			return nil
		},
	}

	sovSigHandler, _ := NewSovereignSubRoundStartExtraSigner(sigHandler, block.OutGoingMbTx)
	err := sovSigHandler.Reset(expectedPubKeys)
	require.Nil(t, err)
	require.True(t, wasResetCalled)
}

func TestSovereignSubRoundStartOutGoingTxData_Identifier(t *testing.T) {
	t.Parallel()

	sovSigHandler, _ := NewSovereignSubRoundStartExtraSigner(&cnsTest.SigningHandlerStub{}, block.OutGoingMbTx)
	require.Equal(t, block.OutGoingMbTx.String(), sovSigHandler.Identifier())
}
