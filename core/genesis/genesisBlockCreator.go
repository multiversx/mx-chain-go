package genesis

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
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

	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if addrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if initialBalances == nil {
		return nil, process.ErrNilValue
	}
	if shardCoordinator == nil {
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
	if addrContainer == nil {
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
