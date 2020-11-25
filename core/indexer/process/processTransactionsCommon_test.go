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
