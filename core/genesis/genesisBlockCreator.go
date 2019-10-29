package genesis

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	factory2 "github.com/ElrondNetwork/elrond-go/vm/factory"
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
func CreateMetaGenesisBlock(
	genesisTime uint64,
	accounts state.AccountsAdapter,
	addrConv state.AddressConverter,
	systemSCs vm.SystemSCContainer,
	txProcessor process.TransactionProcessor,
	nodesSetup *sharding.NodesSetup,
) (data.HeaderHandler, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addrConv == nil || addrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if systemSCs == nil || systemSCs.IsInterfaceNil() {
		return nil, process.ErrNilSystemContractsContainer
	}
	if txProcessor == nil || txProcessor.IsInterfaceNil() {
		return nil, process.ErrNilSmartContractProcessor
	}
	if nodesSetup == nil {
		return nil, process.ErrNilNodesSetup
	}

	err := deploySystemSmartContracts(txProcessor, systemSCs, addrConv)
	if err != nil {
		return nil, err
	}

	err = setStakingData(txProcessor, nodesSetup.InitialNodesInfo(), nodesSetup.StakedValue)
	if err != nil {
		return nil, err
	}

	rootHash, err := accounts.Commit()
	if err != nil {
		return nil, err
	}

	header := &block.MetaBlock{
		RootHash:     rootHash,
		PrevHash:     rootHash,
		RandSeed:     rootHash,
		PrevRandSeed: rootHash,
	}
	header.SetTimeStamp(genesisTime)

	return header, nil
}

// deploySystemSmartContracts deploys all the system smart contracts to the account state
func deploySystemSmartContracts(
	txProcessor process.TransactionProcessor,
	systemSCs vm.SystemSCContainer,
	addrConv state.AddressConverter,
) error {
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(0),
		RcvAddr:   make([]byte, addrConv.AddressLen()),
		GasPrice:  0,
		GasLimit:  0,
		Data:      "deploy@" + hex.EncodeToString(factory.SystemVirtualMachine),
		Signature: nil,
		Challenge: nil,
	}

	for _, key := range systemSCs.Keys() {
		tx.SndAddr = key
		err := txProcessor.ProcessTransaction(tx, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

// setStakingData sets the initial staked values to the staking smart contract
func setStakingData(
	txProcessor process.TransactionProcessor,
	initialNodeInfo map[uint32][]*sharding.NodeInfo,
	stakeValue *big.Int,
) error {
	// create staking smart contract state for genesis - update fixed stake value from all
	for _, nodeInfoList := range initialNodeInfo {
		for _, nodeInfo := range nodeInfoList {
			tx := &transaction.Transaction{
				Nonce:     0,
				Value:     big.NewInt(0).Set(stakeValue),
				RcvAddr:   factory2.StakingSCAddress,
				SndAddr:   nodeInfo.Address(),
				GasPrice:  0,
				GasLimit:  0,
				Data:      "stake@" + hex.EncodeToString(nodeInfo.PubKey()),
				Signature: nil,
				Challenge: nil,
			}

			err := txProcessor.ProcessTransaction(tx, 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
