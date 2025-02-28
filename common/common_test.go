package common

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestIsValidRelayedTxV3(t *testing.T) {
	t.Parallel()

	scr := &smartContractResult.SmartContractResult{}
	require.False(t, IsValidRelayedTxV3(scr))
	require.False(t, IsRelayedTxV3(scr))

	notRelayedTxV3 := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(100),
		RcvAddr:   []byte("receiver"),
		SndAddr:   []byte("sender0"),
		GasPrice:  100,
		GasLimit:  10,
		Signature: []byte("signature"),
	}
	require.False(t, IsValidRelayedTxV3(notRelayedTxV3))
	require.False(t, IsRelayedTxV3(notRelayedTxV3))

	invalidRelayedTxV3 := &transaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(100),
		RcvAddr:     []byte("receiver"),
		SndAddr:     []byte("sender0"),
		GasPrice:    100,
		GasLimit:    10,
		Signature:   []byte("signature"),
		RelayerAddr: []byte("relayer"),
	}
	require.False(t, IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, IsRelayedTxV3(invalidRelayedTxV3))

	invalidRelayedTxV3 = &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(100),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender0"),
		GasPrice:         100,
		GasLimit:         10,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	}
	require.False(t, IsValidRelayedTxV3(invalidRelayedTxV3))
	require.True(t, IsRelayedTxV3(invalidRelayedTxV3))

	relayedTxV3 := &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(100),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender1"),
		GasPrice:         100,
		GasLimit:         10,
		Signature:        []byte("signature"),
		RelayerAddr:      []byte("relayer"),
		RelayerSignature: []byte("signature"),
	}
	require.True(t, IsValidRelayedTxV3(relayedTxV3))
	require.True(t, IsRelayedTxV3(relayedTxV3))
}
