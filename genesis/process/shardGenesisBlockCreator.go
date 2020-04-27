package process

import (
	"fmt"
	"math/big"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("genesis/process")

// CreateShardGenesisBlock will create a shard genesis block
func CreateShardGenesisBlock(arg ArgsGenesisBlockCreator) (data.HeaderHandler, error) {
	rootHash, err := setBalancesToTrie(arg)
	if err != nil {
		return nil, fmt.Errorf("%w encountered when creating genesis block for shard %d",
			err, arg.ShardCoordinator.SelfId())
	}

	//TODO add here delegation process

	header := &block.Header{
		Nonce:           0,
		ShardID:         arg.ShardCoordinator.SelfId(),
		BlockBodyType:   block.StateBlock,
		PubKeysBitmap:   []byte{1},
		Signature:       rootHash,
		RootHash:        rootHash,
		PrevRandSeed:    rootHash,
		RandSeed:        rootHash,
		TimeStamp:       arg.GenesisTime,
		AccumulatedFees: big.NewInt(0),
	}

	return header, nil
}

// setBalancesToTrie adds balances to trie
func setBalancesToTrie(arg ArgsGenesisBlockCreator) (rootHash []byte, err error) {
	if arg.Accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	initialAccounts, err := arg.GenesisParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	initialAccountsForShard := initialAccounts[arg.ShardCoordinator.SelfId()]

	for _, accnt := range initialAccountsForShard {
		err = setBalanceToTrie(arg, accnt)
		if err != nil {
			return nil, err
		}
	}

	rootHash, err = arg.Accounts.Commit()
	if err != nil {
		errToLog := arg.Accounts.RevertToSnapshot(0)
		if errToLog != nil {
			log.Debug("error reverting to snapshot", "error", errToLog.Error())
		}

		return nil, err
	}

	return rootHash, nil
}

func setBalanceToTrie(arg ArgsGenesisBlockCreator, accnt genesis.InitialAccountHandler) error {
	addr, err := arg.PubkeyConv.CreateAddressFromBytes(accnt.AddressBytes())
	if err != nil {
		return fmt.Errorf("%w for address %s", err, accnt.GetAddress())
	}

	accWrp, err := arg.Accounts.LoadAccount(addr)
	if err != nil {
		return err
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = account.AddToBalance(accnt.GetBalanceValue())
	if err != nil {
		return err
	}

	return arg.Accounts.SaveAccount(account)
}
