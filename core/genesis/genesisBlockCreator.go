package genesis

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmFactory "github.com/ElrondNetwork/elrond-go/vm/factory"
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

type ArgsMetaGenesisBlockCreator struct {
	GenesisTime              uint64
	Accounts                 state.AccountsAdapter
	AddrConv                 state.AddressConverter
	NodesSetup               *sharding.NodesSetup
	Economics                *economics.EconomicsData
	ShardCoordinator         sharding.Coordinator
	Store                    dataRetriever.StorageService
	Blkc                     data.ChainHandler
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	MetaDatapool             dataRetriever.MetaPoolsHolder
}

// CreateMetaGenesisBlock creates the meta genesis block
func CreateMetaGenesisBlock(
	args ArgsMetaGenesisBlockCreator,
) (data.HeaderHandler, error) {

	if args.Accounts == nil || args.Accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if args.AddrConv == nil || args.AddrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if args.NodesSetup == nil {
		return nil, process.ErrNilNodesSetup
	}
	if args.Economics == nil {
		return nil, process.ErrNilEconomicsData
	}
	if args.ShardCoordinator == nil || args.ShardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if args.Store == nil || args.Store.IsInterfaceNil() {
		return nil, process.ErrNilStore
	}
	if args.Blkc == nil || args.Blkc.IsInterfaceNil() {
		return nil, process.ErrNilBlockChain
	}
	if args.Marshalizer == nil || args.Marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if args.Hasher == nil || args.Hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if args.Uint64ByteSliceConverter == nil || args.Uint64ByteSliceConverter.IsInterfaceNil() {
		return nil, process.ErrNilUint64Converter
	}
	if args.MetaDatapool == nil || args.MetaDatapool.IsInterfaceNil() {
		return nil, process.ErrNilMetaBlockPool
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         args.Accounts,
		AddrConv:         args.AddrConv,
		StorageService:   args.Store,
		BlockChain:       args.Blkc,
		ShardCoordinator: args.ShardCoordinator,
		Marshalizer:      args.Marshalizer,
		Uint64Converter:  args.Uint64ByteSliceConverter,
	}
	virtualMachineFactory, err := metachain.NewVMContainerFactory(argsHook, args.Economics.StakeValue())
	if err != nil {
		return nil, err
	}

	argsParser, err := smartContract.NewAtArgumentParser()
	if err != nil {
		return nil, err
	}

	vmContainer, err := virtualMachineFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(
		args.ShardCoordinator,
		args.Marshalizer,
		args.Hasher,
		args.AddrConv,
		args.Store,
		args.MetaDatapool,
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(block.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		args.Hasher,
		args.Marshalizer,
		args.Accounts,
		virtualMachineFactory.BlockChainHookImpl(),
		args.AddrConv,
		args.ShardCoordinator,
		scForwarder,
		&metachain.TransactionFeeHandler{},
	)
	if err != nil {
		return nil, err
	}

	txTypeHandler, err := coordinator.NewTxTypeHandler(args.AddrConv, args.ShardCoordinator, args.Accounts)
	if err != nil {
		return nil, err
	}

	txProcessor, err := processTransaction.NewMetaTxProcessor(
		args.Accounts,
		args.AddrConv,
		args.ShardCoordinator,
		scProcessor,
		txTypeHandler,
	)
	if err != nil {
		return nil, process.ErrNilTxProcessor
	}

	err = deploySystemSmartContracts(
		txProcessor,
		virtualMachineFactory.SystemSmartContractContainer(),
		args.AddrConv,
		args.Accounts,
	)
	if err != nil {
		return nil, err
	}

	_, err = args.Accounts.Commit()
	if err != nil {
		return nil, err
	}

	err = setStakingData(txProcessor, args.NodesSetup.InitialNodesInfo(), args.Economics.StakeValue())
	if err != nil {
		return nil, err
	}

	rootHash, err := args.Accounts.Commit()
	if err != nil {
		return nil, err
	}

	header := &block.MetaBlock{
		RootHash:     rootHash,
		PrevHash:     rootHash,
		RandSeed:     rootHash,
		PrevRandSeed: rootHash,
	}
	header.SetTimeStamp(args.GenesisTime)

	return header, nil
}

// deploySystemSmartContracts deploys all the system smart contracts to the account state
func deploySystemSmartContracts(
	txProcessor process.TransactionProcessor,
	systemSCs vm.SystemSCContainer,
	addrConv state.AddressConverter,
	accounts state.AccountsAdapter,
) error {
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(0),
		RcvAddr:   make([]byte, addrConv.AddressLen()),
		GasPrice:  0,
		GasLimit:  0,
		Data:      hex.EncodeToString([]byte("deploy")) + "@" + hex.EncodeToString(factory.SystemVirtualMachine),
		Signature: nil,
		Challenge: nil,
	}

	accountsDB, ok := accounts.(*state.AccountsDB)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	for _, key := range systemSCs.Keys() {
		addr, err := addrConv.CreateAddressFromPublicKeyBytes(key)
		if err != nil {
			return err
		}

		_, err = state.NewAccount(addr, accountsDB)
		if err != nil {
			return err
		}

		_, err = accountsDB.Commit()
		if err != nil {
			return err
		}

		tx.SndAddr = key
		err = txProcessor.ProcessTransaction(tx, 0)
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
				RcvAddr:   vmFactory.StakingSCAddress,
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
