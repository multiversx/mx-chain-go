package process

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory/addressDecoder"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
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

	acc, err = gbc.arg.Accounts.LoadAccount(core.ESDTSCAddress)
	if err != nil {
		return err
	}

	return gbc.arg.Accounts.SaveAccount(acc)
}

func (gbc *sovereignGenesisBlockCreator) createSovereignEmptyGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.computeSovereignDNSAddresses(createSovereignGenesisConfig(gbc.arg.EpochConfig.EnableEpochs))
	if err != nil {
		return nil, err
	}

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(gbc.arg)

	mapEmptyGenesisBlocks := make(map[uint32]data.HeaderHandler, 1)
	mapEmptyGenesisBlocks[core.SovereignChainShardId] = &block.SovereignChainHeader{
		Header: &block.Header{
			Round:     round,
			Nonce:     nonce,
			Epoch:     epoch,
			TimeStamp: gbc.arg.GenesisTime,
			ShardID:   core.SovereignChainShardId,
		},
	}

	return mapEmptyGenesisBlocks, nil
}

func createSovereignGenesisConfig(providedEnableEpochs config.EnableEpochs) config.EnableEpochs {
	cfg := createGenesisConfig(providedEnableEpochs)
	cfg.ESDTMultiTransferEnableEpoch = 0
	return cfg
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

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	allScAddresses := make([][]byte, 0)
	allScAddresses = append(allScAddresses, scResults...)
	genesisBlocks[shardID] = genesisBlock
	err = gbc.saveGenesisBlock(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("'%w' while saving genesis block for shard %d", err, shardID)
	}

	err = gbc.checkDelegationsAgainstDeployedSC(allScAddresses, gbc.arg)
	if err != nil {
		return nil, err
	}

	gb := genesisBlocks[shardID]
	log.Info("sovereignGenesisBlockCreator.createSovereignHeaders",
		"shard", gb.GetShardID(),
		"nonce", gb.GetNonce(),
		"round", gb.GetRound(),
		"root hash", gb.GetRootHash(),
	)

	return genesisBlocks, nil
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
	for _, nodeInfo := range stakedNodes {
		senderAcc, err := arg.Accounts.LoadAccount(nodeInfo.AddressBytes())
		if err != nil {
			return nil, err
		}

		add := hex.EncodeToString(nodeInfo.AddressBytes())
		_ = add

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
	}

	log.Debug("sovereign genesis block",
		"num nodes staked", len(stakedNodes),
	)

	return stakingTxs, nil
}
