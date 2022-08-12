package getAccount

import (
	"math/big"
	"testing"

	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/blockInfoProviders"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAccountsRepository(accDB state.AccountsAdapter, blockchain chainData.ChainHandler) state.AccountsRepository {
	provider, _ := blockInfoProviders.NewCurrentBlockInfo(blockchain)
	wrapper, _ := state.NewAccountsDBApi(accDB, provider)

	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:   wrapper,
		CurrentStateAccountsWrapper: wrapper,
	}
	accountsRepo, _ := state.NewAccountsRepository(args)

	return accountsRepo
}

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	rootHash, _ := accDB.Commit()

	coreComponents := integrationTests.GetDefaultCoreComponents()
	coreComponents.AddressPubKeyConverterField = integrationTests.TestAddressPubkeyConverter

	dataComponents := integrationTests.GetDefaultDataComponents()
	_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: 42}, rootHash)
	dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("header hash"))

	stateComponents := integrationTests.GetDefaultStateComponents()
	stateComponents.AccountsRepo = createAccountsRepository(accDB, dataComponents.BlockChain)

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

	testNonce := uint64(7)
	testBalance := big.NewInt(100)

	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)
	testPubkey := integrationTests.CreateAccount(accDB, testNonce, testBalance)
	rootHash, _ := accDB.Commit()

	coreComponents := integrationTests.GetDefaultCoreComponents()
	coreComponents.AddressPubKeyConverterField = testscommon.RealWorldBech32PubkeyConverter

	dataComponents := integrationTests.GetDefaultDataComponents()
	_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: 42}, rootHash)
	dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("header hash"))

	stateComponents := integrationTests.GetDefaultStateComponents()
	stateComponents.AccountsRepo = createAccountsRepository(accDB, dataComponents.BlockChain)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	testAddress := coreComponents.AddressPubKeyConverter().Encode(testPubkey)
	recovAccnt, _, err := n.GetAccount(testAddress, api.AccountQueryOptions{})

	require.Nil(t, err)
	require.Equal(t, testNonce, recovAccnt.Nonce)
	require.Equal(t, testBalance.String(), recovAccnt.Balance)
}
