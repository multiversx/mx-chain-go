package node

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/stretchr/testify/assert"
)

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	accDB := createInMemoryShardAccountsDB()
	addrConv, _ := addressConverters.NewPlainAddressConverter(32, "")

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressConverter(addrConv),
	)

	recovAccnt, err := n.GetAccount(createDummyHexAddress(64))

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), recovAccnt.Nonce)
	assert.Equal(t, big.NewInt(0), recovAccnt.Balance)
	assert.Nil(t, recovAccnt.CodeHash)
	assert.Nil(t, recovAccnt.RootHash)
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	accDB := createInMemoryShardAccountsDB()
	addrConv, _ := addressConverters.NewPlainAddressConverter(32, "")

	addressHex := createDummyHexAddress(64)
	addressBytes, _ := hex.DecodeString(addressHex)
	address, _ := addrConv.CreateAddressFromPublicKeyBytes(addressBytes)

	nonce := uint64(2233)
	account, _ := accDB.GetAccountWithJournal(address)
	_ = account.SetNonceWithJournal(nonce)

	_, _ = accDB.Commit()

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressConverter(addrConv),
	)

	recovAccnt, err := n.GetAccount(addressHex)

	assert.Nil(t, err)
	assert.Equal(t, nonce, recovAccnt.Nonce)
}
