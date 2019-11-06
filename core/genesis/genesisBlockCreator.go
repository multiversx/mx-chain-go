package genesis

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var log = logger.DefaultLogger()

// CreateShardGenesisBlockFromInitialBalances creates the genesis block body from map of account balances
func CreateShardGenesisBlockFromInitialBalances(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	addrConv state.AddressConverter,
	initialBalances map[string]*big.Int,
	genesisTime uint64,
) (data.HeaderHandler, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addrConv == nil || addrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if initialBalances == nil {
		return nil, process.ErrNilValue
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	rootHash, err := setBalancesToTrie(
		accounts,
		shardCoordinator,
		addrConv,
		initialBalances,
	)
	if err != nil {
		return nil, err
	}

	header := &block.Header{
		Nonce:         0,
		ShardId:       shardCoordinator.SelfId(),
		BlockBodyType: block.StateBlock,
		Signature:     rootHash,
		RootHash:      rootHash,
		PrevRandSeed:  rootHash,
		RandSeed:      rootHash,
		TimeStamp:     genesisTime,
	}

	return header, err
}

// CreateMetaGenesisBlock creates the meta genesis block
func CreateMetaGenesisBlock(genesisTime uint64, _ map[uint32][]string) (data.HeaderHandler, error) {
	//TODO create the right metachain genesis block here
	rootHash := []byte("root hash")
	header := &block.MetaBlock{
		RootHash:     rootHash,
		PrevHash:     rootHash,
		RandSeed:     rootHash,
		PrevRandSeed: rootHash,
	}
	header.SetTimeStamp(genesisTime)

	return header, nil
}

// setBalancesToTrie adds balances to trie
func setBalancesToTrie(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	addrConv state.AddressConverter,
	initialBalances map[string]*big.Int,
) (rootHash []byte, err error) {

	if accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	for i, v := range initialBalances {
		err := setBalanceToTrie(accounts, shardCoordinator, addrConv, []byte(i), v)

		if err != nil {
			return nil, err
		}
	}

	rootHash, err = accounts.Commit()
	if err != nil {
		errToLog := accounts.RevertToSnapshot(0)
		if errToLog != nil {
			log.Error(errToLog.Error())
		}

		return nil, err
	}

	return rootHash, nil
}

func setBalanceToTrie(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	addrConv state.AddressConverter,
	addr []byte,
	balance *big.Int,
) error {

	addrContainer, err := addrConv.CreateAddressFromPublicKeyBytes(addr)
	if err != nil {
		return err
	}
	if addrContainer == nil || addrContainer.IsInterfaceNil() {
		return process.ErrNilAddressContainer
	}
	if shardCoordinator.ComputeId(addrContainer) != shardCoordinator.SelfId() {
		return process.ErrMintAddressNotInThisShard
	}

	accWrp, err := accounts.GetAccountWithJournal(addrContainer)
	if err != nil {
		return err
	}

	account, ok := accWrp.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return account.SetBalanceWithJournal(balance)
}

func initSystemSmartContracts(
	accounts state.AccountsAdapter,
	adrConv state.AddressConverter,
	processor process.SmartContractProcessor,
	smartContracts vm.SystemSCContainer,
) ([]byte, error) {
	scKeys := smartContracts.Keys()

	tx := &transaction.Transaction{
		Value:   big.NewInt(0),
		RcvAddr: make([]byte, adrConv.AddressLen()),
	}
	for _, scKey := range scKeys {
		tx.SndAddr = scKey

		adrSrc, err := adrConv.CreateAddressFromPublicKeyBytes(tx.SndAddr)
		if err != nil {
			return nil, err
		}

		acntSrc, err := accounts.GetAccountWithJournal(adrSrc)
		if err != nil {
			return nil, err
		}

		err = processor.DeploySmartContract(tx, acntSrc, 0)
		if err != nil {
			return nil, err
		}
	}

	return accounts.Commit()
}
