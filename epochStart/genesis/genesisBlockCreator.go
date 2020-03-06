package genesis

import (
	"encoding/hex"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
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
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("core/genesis")

// CreateShardGenesisBlockFromInitialBalances creates the genesis block body from map of account balances
func CreateShardGenesisBlockFromInitialBalances(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	addrConv state.AddressConverter,
	initialBalances map[string]*big.Int,
	genesisTime uint64,
	validatorStatsRootHash []byte,
) (data.HeaderHandler, error) {

	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(addrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if initialBalances == nil {
		return nil, process.ErrNilValue
	}
	if check.IfNil(shardCoordinator) {
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
		Nonce:                  0,
		ShardId:                shardCoordinator.SelfId(),
		BlockBodyType:          block.StateBlock,
		Signature:              rootHash,
		RootHash:               rootHash,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		TimeStamp:              genesisTime,
		ValidatorStatsRootHash: validatorStatsRootHash,
		AccumulatedFees:        big.NewInt(0),
	}

	return header, err
}

// ArgsMetaGenesisBlockCreator holds the arguments which are needed to create a genesis metablock
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
	DataPool                 dataRetriever.PoolsHolder
	ValidatorStatsRootHash   []byte
	MessageSignVerifier      vm.MessageSignVerifier
	GasMap                   map[string]map[string]uint64
}

// CreateMetaGenesisBlock creates the meta genesis block
func CreateMetaGenesisBlock(
	args ArgsMetaGenesisBlockCreator,
) (data.HeaderHandler, error) {

	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.AddrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if args.NodesSetup == nil {
		return nil, process.ErrNilNodesSetup
	}
	if args.Economics == nil {
		return nil, process.ErrNilEconomicsData
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(args.Blkc) {
		return nil, process.ErrNilBlockChain
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return nil, process.ErrNilUint64Converter
	}
	if check.IfNil(args.DataPool) {
		return nil, process.ErrNilMetaBlocksPool
	}

	txProcessor, systemSmartContracts, err := createProcessorsForMetaGenesisBlock(args)
	if err != nil {
		return nil, err
	}

	err = deploySystemSmartContracts(
		txProcessor,
		systemSmartContracts,
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

	eligible, waiting := args.NodesSetup.InitialNodesInfo()
	allNodes := make(map[uint32][]*sharding.NodeInfo)

	for shard := range eligible {
		allNodes[shard] = append(eligible[shard], waiting[shard]...)
	}

	err = setStakedData(
		txProcessor,
		allNodes,
		args.Economics.GenesisNodePrice(),
	)
	if err != nil {
		return nil, err
	}

	rootHash, err := args.Accounts.Commit()
	if err != nil {
		return nil, err
	}

	header := &block.MetaBlock{
		RootHash:               rootHash,
		PrevHash:               rootHash,
		RandSeed:               rootHash,
		PrevRandSeed:           rootHash,
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
	}
	header.EpochStart.Economics = block.Economics{
		TotalSupply:            big.NewInt(0).Set(args.Economics.GenesisTotalSupply()),
		TotalToDistribute:      big.NewInt(0),
		TotalNewlyMinted:       big.NewInt(0),
		RewardsPerBlockPerNode: big.NewInt(0),
		NodePrice:              big.NewInt(0).Set(args.Economics.GenesisNodePrice()),
	}

	header.SetTimeStamp(args.GenesisTime)
	header.SetValidatorStatsRootHash(args.ValidatorStatsRootHash)

	err = saveGenesisMetaToStorage(args.Store, args.Marshalizer, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func saveGenesisMetaToStorage(
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	genesisBlock data.HeaderHandler,
) error {

	epochStartID := core.EpochStartIdentifier(0)
	metaHdrStorage := storageService.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(metaHdrStorage) {
		return epochStart.ErrNilStorage
	}

	marshaledData, err := marshalizer.Marshal(genesisBlock)
	if err != nil {
		return err
	}

	err = metaHdrStorage.Put([]byte(epochStartID), marshaledData)
	if err != nil {
		return err
	}

	return nil
}

func createProcessorsForMetaGenesisBlock(
	args ArgsMetaGenesisBlockCreator,
) (process.TransactionProcessor, vm.SystemSCContainer, error) {
	argsHook := hooks.ArgBlockChainHook{
		Accounts:         args.Accounts,
		AddrConv:         args.AddrConv,
		StorageService:   args.Store,
		BlockChain:       args.Blkc,
		ShardCoordinator: args.ShardCoordinator,
		Marshalizer:      args.Marshalizer,
		Uint64Converter:  args.Uint64ByteSliceConverter,
	}

	virtualMachineFactory, err := metachain.NewVMContainerFactory(argsHook, args.Economics, &NilMessageSignVerifier{}, args.GasMap)
	if err != nil {
		return nil, nil, err
	}

	argsParser := vmcommon.NewAtArgumentParser()

	vmContainer, err := virtualMachineFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(
		args.ShardCoordinator,
		args.Marshalizer,
		args.Hasher,
		args.AddrConv,
		args.Store,
		args.DataPool,
	)
	if err != nil {
		return nil, nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	scForwarder, err := interimProcContainer.Get(block.SmartContractResultBlock)
	if err != nil {
		return nil, nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(args.Economics)
	if err != nil {
		return nil, nil, err
	}

	txTypeHandler, err := coordinator.NewTxTypeHandler(args.AddrConv, args.ShardCoordinator, args.Accounts)
	if err != nil {
		return nil, nil, err
	}

	genesisFeeHandler := NewGenesisFeeHandler()
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:   vmContainer,
		ArgsParser:    argsParser,
		Hasher:        args.Hasher,
		Marshalizer:   args.Marshalizer,
		AccountsDB:    args.Accounts,
		TempAccounts:  virtualMachineFactory.BlockChainHookImpl(),
		AdrConv:       args.AddrConv,
		Coordinator:   args.ShardCoordinator,
		ScrForwarder:  scForwarder,
		TxFeeHandler:  genesisFeeHandler,
		EconomicsFee:  genesisFeeHandler,
		TxTypeHandler: txTypeHandler,
		GasHandler:    gasHandler,
		GasMap:        args.GasMap,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewSCProcessor)
	if err != nil {
		return nil, nil, err
	}

	txProcessor, err := processTransaction.NewMetaTxProcessor(
		args.Accounts,
		args.AddrConv,
		args.ShardCoordinator,
		scProcessor,
		txTypeHandler,
		genesisFeeHandler,
	)
	if err != nil {
		return nil, nil, process.ErrNilTxProcessor
	}

	return txProcessor, virtualMachineFactory.SystemSmartContractContainer(), nil
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
		GasLimit:  math.MaxUint64,
		Data:      []byte(hex.EncodeToString([]byte("deploy")) + "@" + hex.EncodeToString(factory.SystemVirtualMachine)),
		Signature: nil,
	}

	accountsDB, ok := accounts.(*state.AccountsDB)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	for _, key := range systemSCs.Keys() {
		//addr, err := addrConv.CreateAddressFromPublicKeyBytes(key)
		//if err != nil {
		//	return err
		//}

		//_, err = state.NewAccount(addr, accountsDB)
		//if err != nil {
		//	return err
		//}

		_, err := accountsDB.Commit()
		if err != nil {
			return err
		}

		tx.SndAddr = key
		err = txProcessor.ProcessTransaction(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

// setStakedData sets the initial staked values to the staking smart contract
func setStakedData(
	txProcessor process.TransactionProcessor,
	initialNodeInfo map[uint32][]*sharding.NodeInfo,
	stakeValue *big.Int,
) error {
	// create staking smart contract state for genesis - update fixed stake value from all
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	for i := range initialNodeInfo {
		nodeInfoList := initialNodeInfo[i]
		for _, nodeInfo := range nodeInfoList {
			tx := &transaction.Transaction{
				Nonce:     0,
				Value:     big.NewInt(0).Set(stakeValue),
				RcvAddr:   vmFactory.AuctionSCAddress,
				SndAddr:   nodeInfo.Address(),
				GasPrice:  0,
				GasLimit:  math.MaxUint64,
				Data:      []byte("stake@" + oneEncoded + "@" + hex.EncodeToString(nodeInfo.PubKey()) + "@" + hex.EncodeToString([]byte("genesis"))),
				Signature: nil,
			}

			err := txProcessor.ProcessTransaction(tx)
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
		err = setBalanceToTrie(accounts, shardCoordinator, addrConv, []byte(i), v)

		if err != nil {
			return nil, err
		}
	}

	rootHash, err = accounts.Commit()
	if err != nil {
		errToLog := accounts.RevertToSnapshot(0)
		if errToLog != nil {
			log.Debug("error reverting to snapshot", "error", errToLog.Error())
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

	accWrp, err := accounts.LoadAccount(addrContainer)
	if err != nil {
		return err
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = account.AddToBalance(balance)
	if err != nil {
		return err
	}

	return accounts.SaveAccount(account)
}

type NilMessageSignVerifier struct {
}

func (n *NilMessageSignVerifier) Verify(_ []byte, _ []byte, _ []byte) error {
	return nil
}

func (n *NilMessageSignVerifier) IsInterfaceNil() bool {
	return n == nil
}
