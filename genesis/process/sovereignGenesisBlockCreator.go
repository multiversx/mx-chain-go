package process

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/factory/addressDecoder"
	"github.com/multiversx/mx-chain-go/genesis"
	genesisCommon "github.com/multiversx/mx-chain-go/genesis/process/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
)

type sovereignGenesisBlockCreator struct {
	*genesisBlockCreator
}

// NewSovereignGenesisBlockCreator creates a new sovereign genesis block creator instance
func NewSovereignGenesisBlockCreator(gbc *genesisBlockCreator) (*sovereignGenesisBlockCreator, error) {
	if gbc == nil {
		return nil, errNilGenesisBlockCreator
	}

	log.Debug("NewSovereignGenesisBlockCreator", "native esdt token", gbc.arg.Config.SovereignConfig.GenesisConfig.NativeESDT)

	return &sovereignGenesisBlockCreator{
		genesisBlockCreator: gbc,
	}, nil
}

// CreateGenesisBlocks will create sovereign genesis blocks
func (gbc *sovereignGenesisBlockCreator) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.initGenesisAccounts()
	if err != nil {
		return nil, err
	}
	if !mustDoGenesisProcess(gbc.arg) {
		return gbc.createSovereignEmptyGenesisBlocks()
	}

	err = gbc.computeSovereignDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
	if err != nil {
		return nil, err
	}

	shardIDs := make([]uint32, 1)
	shardIDs[0] = core.SovereignChainShardId
	argsCreateBlock, err := gbc.createGenesisBlocksArgs(shardIDs)
	if err != nil {
		return nil, err
	}

	return gbc.createSovereignHeaders(argsCreateBlock)
}

func (gbc *sovereignGenesisBlockCreator) initGenesisAccounts() error {
	acc, err := gbc.arg.Accounts.LoadAccount(core.SystemAccountAddress)
	if err != nil {
		return err
	}

	err = gbc.arg.Accounts.SaveAccount(acc)
	if err != nil {
		return err
	}

	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	return genesisCommon.UpdateSystemSCContractsCode(codeMetaData.ToBytes(), gbc.arg.Accounts)
}

func (gbc *sovereignGenesisBlockCreator) createSovereignEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeSovereignDNSAddresses(createSovereignGenesisConfig(gbc.arg.EpochConfig.EnableEpochs))
	if err != nil {
		return nil, err
	}

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(gbc.arg)

	genesisEmptyBlock := gbc.arg.RunTypeComponents.VersionedHeaderFactory().Create(epoch)
	err = gbc.setGenesisEmptyBlockData(genesisEmptyBlock, round, nonce)
	if err != nil {
		return nil, err
	}

	mapEmptyGenesisBlocks := make(map[uint32]data.HeaderHandler, 1)
	mapEmptyGenesisBlocks[core.SovereignChainShardId] = genesisEmptyBlock
	return mapEmptyGenesisBlocks, nil
}

func (gbc *sovereignGenesisBlockCreator) setGenesisEmptyBlockData(
	genesisEmptyBlock data.HeaderHandler,
	round uint64,
	nonce uint64,
) error {
	err := genesisEmptyBlock.SetShardID(core.SovereignChainShardId)
	if err != nil {
		return err
	}
	err = genesisEmptyBlock.SetRound(round)
	if err != nil {
		return err
	}
	err = genesisEmptyBlock.SetNonce(nonce)
	if err != nil {
		return err
	}
	return genesisEmptyBlock.SetTimeStamp(gbc.arg.GenesisTime)
}

func createSovereignGenesisConfig(providedEnableEpochs config.EnableEpochs) config.EnableEpochs {
	return providedEnableEpochs
}

func (gbc *sovereignGenesisBlockCreator) computeSovereignDNSAddresses(enableEpochsConfig config.EnableEpochs) error {
	initialAddresses, err := addressDecoder.DecodeAddresses(gbc.arg.Core.AddressPubKeyConverter(), gbc.arg.DNSV2Addresses)
	if err != nil {
		return err
	}

	return gbc.computeDNSAddresses(enableEpochsConfig, initialAddresses)
}

func (gbc *sovereignGenesisBlockCreator) createSovereignHeaders(args *headerCreatorArgs) (map[uint32]data.HeaderHandler, error) {
	shardID := core.SovereignChainShardId
	log.Debug("sovereignGenesisBlockCreator.createSovereignHeaders", "shard", shardID)

	var genesisBlock data.HeaderHandler
	var scResults [][]byte
	var err error

	genesisBlock, scResults, gbc.initialIndexingData[shardID], err = createSovereignShardGenesisBlock(
		args.mapArgsGenesisBlockCreator[shardID],
		args.nodesListSplitter,
	)

	if err != nil {
		return nil, fmt.Errorf("'%w' while generating genesis block for shard %d", err, shardID)
	}

	err = gbc.saveGenesisBlock(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("'%w' while saving genesis block for shard %d", err, shardID)
	}

	err = gbc.checkDelegationsAgainstDeployedSC(scResults, gbc.arg)
	if err != nil {
		return nil, err
	}

	err = genesisBlock.SetSoftwareVersion(process.SovereignHeaderVersion)
	if err != nil {
		return nil, err
	}

	prevHash := gbc.arg.Core.Hasher().Compute(gbc.arg.GenesisString)
	err = genesisBlock.SetPrevHash(prevHash)
	if err != nil {
		return nil, err
	}

	validatorRootHash, err := gbc.arg.ValidatorAccounts.RootHash()
	if err != nil {
		return nil, err
	}

	err = gbc.setInitialDataInSovereignHeader(genesisBlock, validatorRootHash)
	if err != nil {
		return nil, err
	}

	err = saveSovereignGenesisStorage(gbc.arg.Data.StorageService(), gbc.arg.Core.InternalMarshalizer(), genesisBlock)
	if err != nil {
		return nil, err
	}

	log.Info("sovereignGenesisBlockCreator.createSovereignHeaders",
		"shard", genesisBlock.GetShardID(),
		"nonce", genesisBlock.GetNonce(),
		"round", genesisBlock.GetRound(),
		"root hash", genesisBlock.GetRootHash(),
	)

	return map[uint32]data.HeaderHandler{
		core.SovereignChainShardId: genesisBlock,
	}, nil
}

func (gbc *sovereignGenesisBlockCreator) setInitialDataInSovereignHeader(header data.HeaderHandler, validatorRootHash []byte) error {
	sovereignHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return process.ErrWrongTypeAssertion
	}

	zeroBI := big.NewInt(0)
	err := sovereignHeader.SetAccumulatedFeesInEpoch(zeroBI)
	if err != nil {
		return err
	}
	err = sovereignHeader.SetDevFeesInEpoch(zeroBI)
	if err != nil {
		return err
	}
	err = sovereignHeader.SetValidatorStatsRootHash(validatorRootHash)
	if err != nil {
		return err
	}
	err = sovereignHeader.SetStartOfEpochHeader()
	if err != nil {
		return err
	}
	return sovereignHeader.GetEpochStartHandler().SetEconomics(&block.Economics{
		TotalSupply:       big.NewInt(0).Set(gbc.arg.Economics.GenesisTotalSupply()),
		TotalToDistribute: big.NewInt(0),
		TotalNewlyMinted:  big.NewInt(0),
		RewardsPerBlock:   big.NewInt(0),
		NodePrice:         big.NewInt(0).Set(gbc.arg.GenesisNodePrice),
	})
}

func saveSovereignGenesisStorage(
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	genesisBlock data.HeaderHandler,
) error {
	epochStartID := core.EpochStartIdentifier(genesisBlock.GetEpoch())

	blockHdrStorage, err := storageService.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return err
	}

	triggerStorage, err := storageService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return err
	}

	marshaledData, err := marshalizer.Marshal(genesisBlock)
	if err != nil {
		return err
	}

	err = blockHdrStorage.Put([]byte(epochStartID), marshaledData)
	if err != nil {
		return err
	}

	return triggerStorage.Put([]byte(epochStartID), marshaledData)
}

func createSovereignShardGenesisBlock(
	arg ArgsGenesisBlockCreator,
	nodesListSplitter genesis.NodesListSplitter,
) (data.HeaderHandler, [][]byte, *genesis.IndexingData, error) {
	sovereignGenesisConfig := createSovereignGenesisConfig(arg.EpochConfig.EnableEpochs)
	shardProcessors, err := createProcessorsForShardGenesisBlock(arg, sovereignGenesisConfig, createGenesisRoundConfig(arg.RoundConfig))
	if err != nil {
		return nil, nil, nil, err
	}

	genesisBlock, scAddresses, indexingData, err := baseCreateShardGenesisBlock(arg, nodesListSplitter, shardProcessors)
	if err != nil {
		return nil, nil, nil, err
	}

	metaProcessor, err := createProcessorsForMetaGenesisBlock(arg, sovereignGenesisConfig, createGenesisRoundConfig(arg.RoundConfig))
	if err != nil {
		return nil, nil, nil, err
	}

	err = initSystemSCs(shardProcessors.vmContainer, arg.Accounts)
	if err != nil {
		return nil, nil, nil, err
	}

	deploySystemSCTxs, err := deploySystemSmartContracts(arg, metaProcessor.txProcessor, metaProcessor.systemSCs)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.DeploySystemScTxs = deploySystemSCTxs

	stakingTxs, err := setSovereignStakedData(arg, metaProcessor, nodesListSplitter)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.StakingTxs = stakingTxs

	metaScrsTxs := metaProcessor.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	indexingData.ScrsTxs = mergeScrs(indexingData.ScrsTxs, metaScrsTxs)

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating sovereign genesis block while commiting", err)
	}

	err = setRootHash(genesisBlock, rootHash)
	if err != nil {
		return nil, nil, nil, err
	}

	validatorRootHash, err := arg.ValidatorAccounts.RootHash()
	if err != nil {
		return nil, nil, nil, err
	}

	err = genesisBlock.SetValidatorStatsRootHash(validatorRootHash)
	if err != nil {
		return nil, nil, nil, err
	}

	err = metaProcessor.vmContainer.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	return genesisBlock, scAddresses, indexingData, nil
}

func initSystemSCs(vmContainer process.VirtualMachinesContainer, accounts state.AccountsAdapter) error {
	vmExecutionHandler, err := vmContainer.Get(factory.SystemVirtualMachine)
	if err != nil {
		return err
	}

	err = genesisCommon.InitDelegationSystemSC(vmExecutionHandler, accounts)
	if err != nil {
		return err
	}

	return genesisCommon.InitGovernanceV2(vmExecutionHandler, accounts)
}

func setRootHash(header data.HeaderHandler, rootHash []byte) error {
	err := header.SetSignature(rootHash)
	if err != nil {
		return err
	}
	err = header.SetRootHash(rootHash)
	if err != nil {
		return err
	}
	err = header.SetPrevRandSeed(rootHash)
	if err != nil {
		return err
	}

	return header.SetRandSeed(rootHash)
}

func mergeScrs(scrTxs1, scrTxs2 map[string]data.TransactionHandler) map[string]data.TransactionHandler {
	allScrsTxs := make(map[string]data.TransactionHandler)
	for scrTxIn1, scrTx := range scrTxs1 {
		allScrsTxs[scrTxIn1] = scrTx
	}

	for scrTxIn2, scrTx := range scrTxs2 {
		allScrsTxs[scrTxIn2] = scrTx
	}

	return allScrsTxs
}

func setSovereignStakedData(
	arg ArgsGenesisBlockCreator,
	processors *genesisProcessors,
	nodesListSplitter genesis.NodesListSplitter,
) ([]data.TransactionHandler, error) {
	scQueryBlsKeys := &process.SCQuery{
		ScAddress: vm.StakingSCAddress,
		FuncName:  "isStaked",
	}

	stakingTxs := make([]data.TransactionHandler, 0)

	// create staking smart contract state for genesis - update fixed stake value from all
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	stakeValue := arg.GenesisNodePrice

	stakedNodes := nodesListSplitter.GetAllNodes()
	argsUpdateOwnersForBlsKeys := make([][]byte, 0)
	for idx, nodeInfo := range stakedNodes {
		senderAcc, err := arg.Accounts.LoadAccount(nodeInfo.AddressBytes())
		if err != nil {
			return nil, err
		}

		tx := &transaction.Transaction{
			Nonce:     senderAcc.GetNonce(),
			Value:     new(big.Int).Set(stakeValue),
			RcvAddr:   vm.ValidatorSCAddress,
			SndAddr:   nodeInfo.AddressBytes(),
			GasPrice:  0,
			GasLimit:  math.MaxUint64,
			Data:      []byte("stake@" + oneEncoded + "@" + hex.EncodeToString(nodeInfo.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis"))),
			Signature: nil,
		}

		stakingTxs = append(stakingTxs, tx)

		_, err = processors.txProcessor.ProcessTransaction(tx)
		if err != nil {
			return nil, err
		}

		scQueryBlsKeys.Arguments = [][]byte{nodeInfo.PubKeyBytes()}
		vmOutput, _, err := processors.queryService.ExecuteQuery(scQueryBlsKeys)
		if err != nil {
			return nil, err
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			return nil, genesis.ErrBLSKeyNotStaked
		}

		err = setGenesisNodeChainID(idx, arg.ValidatorAccounts, nodeInfo.PubKeyBytes())
		if err != nil {
			return nil, err
		}

		argsUpdateOwnersForBlsKeys = append(argsUpdateOwnersForBlsKeys, nodeInfo.PubKeyBytes())
		argsUpdateOwnersForBlsKeys = append(argsUpdateOwnersForBlsKeys, senderAcc.AddressBytes())
	}

	err := updateOwnersForBlsKeys(argsUpdateOwnersForBlsKeys, processors.vmContainer, arg.Accounts)
	if err != nil {
		return nil, err
	}

	log.Debug("sovereign genesis block",
		"num nodes staked", len(stakedNodes),
	)

	return stakingTxs, nil
}

// setGenesisNodeChainID assigns ascending numerical IDs to BLS keys, based on their order in the genesis config file.
// Each ID is stored as a 2-byte big-endian number.
func setGenesisNodeChainID(id int, peerAccountsDB state.AccountsAdapter, key []byte) error {
	valAcc, err := getPeerAccount(peerAccountsDB, key)
	if err != nil {
		return err
	}

	valAcc.SetMainChainID(intTo2Bytes(id))
	return peerAccountsDB.SaveAccount(valAcc)
}

func intTo2Bytes(n int) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(n)) // Convert only the lower 2 bytes
	return b
}

func getPeerAccount(peerAccountsDB state.AccountsAdapter, key []byte) (state.PeerAccountHandler, error) {
	account, err := peerAccountsDB.LoadAccount(key)
	if err != nil {
		return nil, err
	}

	peerAcc, ok := account.(state.PeerAccountHandler)
	if !ok {
		return nil, epochStart.ErrWrongTypeAssertion
	}

	return peerAcc, nil
}

func updateOwnersForBlsKeys(
	arguments [][]byte,
	vmContainer process.VirtualMachinesContainer,
	accounts state.AccountsAdapter,
) error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.EndOfEpochAddress,
			CallValue:  big.NewInt(0),
			Arguments:  arguments,
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "setOwnersOnAddresses",
	}

	systemVM, err := vmContainer.Get(factory.SystemVirtualMachine)
	if err != nil {
		return err
	}

	vmOutput, errRun := systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return fmt.Errorf("%w when calling setOwnersOnAddresses function", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return fmt.Errorf("got return code %s when calling setOwnersOnAddresses", vmOutput.ReturnCode)
	}

	return genesisCommon.ProcessSCOutputAccounts(vmOutput, accounts)
}
