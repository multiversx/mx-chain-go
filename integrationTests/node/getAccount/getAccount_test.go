package getAccount

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/stretchr/testify/assert"
)

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	coreComponents := integrationTests.GetDefaultCoreComponents()
	coreComponents.AddressPubKeyConverterField = integrationTests.TestAddressPubkeyConverter

	stateComponents := integrationTests.GetDefaultStateComponents()
	stateComponents.AccountsAPI = accDB

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	encodedAddress := integrationTests.TestAddressPubkeyConverter.Encode(integrationTests.CreateRandomBytes(32))
	recovAccnt, err := n.GetAccount(encodedAddress)

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), recovAccnt.Nonce)
	assert.Equal(t, "0", recovAccnt.Balance)
	assert.Equal(t, "0", recovAccnt.DeveloperReward)
	assert.Empty(t, recovAccnt.OwnerAddress)
	assert.Nil(t, recovAccnt.CodeHash)
	assert.Nil(t, recovAccnt.RootHash)
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	addressBytes := integrationTests.CreateRandomBytes(32)
	nonce := uint64(2233)
	account, _ := accDB.LoadAccount(addressBytes)
	account.IncreaseNonce(nonce)
	_ = accDB.SaveAccount(account)
	_, _ = accDB.Commit()

	coreComponents := integrationTests.GetDefaultCoreComponents()
	coreComponents.AddressPubKeyConverterField = integrationTests.TestAddressPubkeyConverter

	stateComponents := integrationTests.GetDefaultStateComponents()
	stateComponents.AccountsAPI = accDB

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	encodedAddress := integrationTests.TestAddressPubkeyConverter.Encode(addressBytes)
	recovAccnt, err := n.GetAccount(encodedAddress)

	assert.Nil(t, err)
	assert.Equal(t, nonce, recovAccnt.Nonce)
}
