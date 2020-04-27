package process

import (
	"fmt"
	"math/big"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

func setBalanceToTrie(arg ArgsGenesisBlockCreator, accnt *genesis.InitialAccount) error {
	addr, err := arg.PubkeyConv.CreateAddressFromBytes(accnt.AddressBytes)
	if err != nil {
		return fmt.Errorf("%w for address %s", err, accnt.Address)
	}

	accWrp, err := arg.Accounts.LoadAccount(addr)
	if err != nil {
		return err
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = account.AddToBalance(accnt.Balance)
	if err != nil {
		return err
	}

	return arg.Accounts.SaveAccount(account)
}

func createProcessorsForShard(arg ArgsGenesisBlockCreator) (*genesisProcessors, error) {
	argsParser := vmcommon.NewAtArgumentParser()

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          gasSchedule,
		MapDNSAddresses: make(map[string]struct{}),
	}
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         stateComponents.AccountsAdapter,
		PubkeyConv:       stateComponents.AddressPubkeyConverter,
		StorageService:   data.Store,
		BlockChain:       data.Blkc,
		ShardCoordinator: shardCoordinator,
		Marshalizer:      core.InternalMarshalizer,
		Uint64Converter:  core.Uint64ByteSliceConverter,
		BuiltInFunctions: builtInFuncs,
	}
	vmFactory, err := shard.NewVMContainerFactory(config.VirtualMachineConfig, economics.MaxGasLimitPerBlock(), gasSchedule, argsHook)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		core.InternalMarshalizer,
		core.Hasher,
		stateComponents.AddressPubkeyConverter,
		data.Store,
		data.Datapool,
		economics,
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	receiptTxInterim, err := interimProcContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return nil, err
	}

	badTxInterim, err := interimProcContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(economics)
	if err != nil {
		return nil, err
	}

	txFeeHandler, err := postprocess.NewFeeAccumulator()
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  stateComponents.AddressPubkeyConverter,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   vmcommon.NewAtArgumentParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       argsParser,
		Hasher:           core.Hasher,
		Marshalizer:      core.InternalMarshalizer,
		AccountsDB:       stateComponents.AccountsAdapter,
		TempAccounts:     vmFactory.BlockChainHookImpl(),
		PubkeyConv:       stateComponents.AddressPubkeyConverter,
		Coordinator:      shardCoordinator,
		ScrForwarder:     scForwarder,
		TxFeeHandler:     txFeeHandler,
		EconomicsFee:     economics,
		TxTypeHandler:    txTypeHandler,
		GasHandler:       gasHandler,
		BuiltInFunctions: vmFactory.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:  txLogsProcessor,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		stateComponents.AccountsAdapter,
		stateComponents.AddressPubkeyConverter,
		shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewTxProcessor(
		stateComponents.AccountsAdapter,
		core.Hasher,
		stateComponents.AddressPubkeyConverter,
		core.InternalMarshalizer,
		shardCoordinator,
		scProcessor,
		txFeeHandler,
		txTypeHandler,
		economics,
		receiptTxInterim,
		badTxInterim,
	)
	if err != nil {
		return nil, errors.New("could not create transaction statisticsProcessor: " + err.Error())
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	blockSizeComputationHandler, err := preprocess.NewBlockSizeComputation(core.InternalMarshalizer, blockSizeThrottler, maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	balanceComputationHandler, err := preprocess.NewBalanceComputation()
	if err != nil {
		return nil, err
	}

	preProcFactory, err := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		data.Store,
		core.InternalMarshalizer,
		core.Hasher,
		data.Datapool,
		stateComponents.AddressPubkeyConverter,
		stateComponents.AccountsAdapter,
		requestHandler,
		transactionProcessor,
		scProcessor,
		scProcessor,
		rewardsTxProcessor,
		economics,
		gasHandler,
		blockTracker,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		core.Hasher,
		core.InternalMarshalizer,
		shardCoordinator,
		stateComponents.AccountsAdapter,
		data.Datapool.MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		txFeeHandler,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
