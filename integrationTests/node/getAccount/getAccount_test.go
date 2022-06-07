package getAccount

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	// Question for review >> how to store data in the trieStorage for a test root hash,
	// so that recreateTrie() doesn't fail for []byte("root hash")?
	// Alternatively, I've altered this test (as a temporary workarounf)
	// << question notes review

	// trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	// accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	accountsRepository := testscommon.NewAccountsRepositoryStub()
	accountsRepository.GetExistingAccountCalled = func(address, rootHash []byte) (vmcommon.AccountHandler, error) {
		return nil, state.ErrAccNotFound
	}

	coreComponents := integrationTests.GetDefaultCoreComponents()
	coreComponents.AddressPubKeyConverterField = integrationTests.TestAddressPubkeyConverter

	dataComponents := integrationTests.GetDefaultDataComponents()
	_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: 42}, []byte("root hash"))
	dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("header hash"))

	stateComponents := integrationTests.GetDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	encodedAddress := integrationTests.TestAddressPubkeyConverter.Encode(integrationTests.CreateRandomBytes(32))
	recovAccnt, _, err := n.GetAccount(encodedAddress, api.AccountQueryOptions{})

	require.Nil(t, err)
	assert.Equal(t, uint64(0), recovAccnt.Nonce)
	assert.Equal(t, "0", recovAccnt.Balance)
	assert.Equal(t, "0", recovAccnt.DeveloperReward)
	assert.Empty(t, recovAccnt.OwnerAddress)
	assert.Nil(t, recovAccnt.CodeHash)
	assert.Nil(t, recovAccnt.RootHash)
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	// Question for review: same question as above
	// trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	// accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	accountsRepository := testscommon.NewAccountsRepositoryStub()
	testNonce := uint64(2233)
	testAccount, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	testAccount.Nonce = testNonce
	accountsRepository.AddTestAccount(testAccount)

	coreComponents := integrationTests.GetDefaultCoreComponents()
	coreComponents.AddressPubKeyConverterField = testscommon.RealWorldBech32PubkeyConverter

	dataComponents := integrationTests.GetDefaultDataComponents()
	_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: 42}, []byte("root hash"))
	dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("header hash"))

	stateComponents := integrationTests.GetDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	recovAccnt, _, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})

	assert.Nil(t, err)
	assert.Equal(t, testNonce, recovAccnt.Nonce)
}
