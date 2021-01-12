package process

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestAddToAlteredAddresses(t *testing.T) {
	t.Parallel()

	sender := "senderAddress"
	receiver := "receiverAddress"
	tokenIdentifier := "Test-token"
	tx := &types.Transaction{
		Sender:              sender,
		Receiver:            receiver,
		EsdtValue:           "123",
		EsdtTokenIdentifier: tokenIdentifier,
	}
	alteredAddress := make(map[string]*accounts.AlteredAccount)
	selfShardID := uint32(0)
	mb := &block.MiniBlock{}

	addToAlteredAddresses(tx, alteredAddress, mb, selfShardID, false)

	alterdAccount, ok := alteredAddress[receiver]
	require.True(t, ok)
	require.Equal(t, &accounts.AlteredAccount{
		IsESDTSender:    false,
		IsESDTOperation: true,
		TokenIdentifier: tokenIdentifier,
	}, alterdAccount)

	alterdAccount, ok = alteredAddress[sender]
	require.True(t, ok)
	require.Equal(t, &accounts.AlteredAccount{
		IsESDTSender:    true,
		IsESDTOperation: true,
		TokenIdentifier: tokenIdentifier,
	}, alterdAccount)
}

func TestIsSCRForSenderWithGasUsed(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	nonce := uint64(10)
	sender := "sender"

	tx := &types.Transaction{
		Hash:   txHash,
		Nonce:  nonce,
		Sender: sender,
	}
	sc := &types.ScResult{
		Data:      []byte("@6f6b@something"),
		Nonce:     nonce + 1,
		Receiver:  sender,
		PreTxHash: txHash,
	}

	require.True(t, isSCRForSenderWithRefund(sc, tx))
}

func TestComputeTxGasUsedField(t *testing.T) {
	t.Parallel()

	tx := &types.Transaction{
		GasLimit: 500,
		GasPrice: 10,
	}
	sc := &types.ScResult{
		Value: "3000",
	}

	expectedGasUsed := uint64(200)
	require.Equal(t, expectedGasUsed, computeTxGasUsedField(sc, tx))
}
