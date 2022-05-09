package staking

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

type OwnerStats struct {
	EligibleBlsKeys  map[uint32][][]byte
	WaitingBlsKeys   map[uint32][][]byte
	StakingQueueKeys [][]byte
	TotalStake       *big.Int
}

type InitialNodesConfig struct {
	Owners                        map[string]*OwnerStats
	MaxNodesChangeConfig          []config.MaxNodesChangeConfig
	NumOfShards                   uint32
	MinNumberOfEligibleShardNodes uint32
	MinNumberOfEligibleMetaNodes  uint32
	ShardConsensusGroupSize       int
	MetaConsensusGroupSize        int
}

func NewTestMetaProcessorWithCustomNodes(config *InitialNodesConfig) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(config.NumOfShards)

	_ = dataComponents
	_ = bootstrapComponents
	_ = statusComponents

	queue := createStakingQueueCustomNodes(
		config.Owners,
		coreComponents.InternalMarshalizer(),
		stateComponents.AccountsAdapter(),
	)

	eligibleMap, waitingMap := createGenesisNodesWithCustomConfig(
		config.Owners,
		coreComponents.InternalMarshalizer(),
		stateComponents,
	)

	nc := createNodesCoordinator(
		eligibleMap,
		waitingMap,
		config.MinNumberOfEligibleMetaNodes,
		config.NumOfShards,
		config.MinNumberOfEligibleShardNodes,
		config.ShardConsensusGroupSize,
		config.MetaConsensusGroupSize,
		coreComponents,
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		bootstrapComponents.NodesCoordinatorRegistryFactory(),
		config.MaxNodesChangeConfig,
	)

	return newTestMetaProcessor(
		coreComponents,
		dataComponents,
		bootstrapComponents,
		statusComponents,
		stateComponents,
		nc,
		config.MaxNodesChangeConfig,
		queue,
	)
}

type NodesRegisterData struct {
	BLSKeys    [][]byte
	TotalStake *big.Int
}

func (tmp *TestMetaProcessor) ProcessStake(t *testing.T, nodes map[string]*NodesRegisterData) {
	header := tmp.createNewHeader(t, tmp.currentRound)
	tmp.BlockChainHook.SetCurrentHeader(header)

	txHashes := make([][]byte, 0)
	for owner, nodesData := range nodes {
		numBLSKeys := int64(len(nodesData.BLSKeys))
		numBLSKeysBytes := big.NewInt(numBLSKeys).Bytes()

		txData := hex.EncodeToString([]byte("stake")) + "@" + hex.EncodeToString(numBLSKeysBytes)
		argsStake := [][]byte{numBLSKeysBytes}

		for _, blsKey := range nodesData.BLSKeys {
			signature := append([]byte("signature-"), blsKey...)

			argsStake = append(argsStake, blsKey, signature)
			txData += "@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(signature)
		}

		txHash := append([]byte("txHash-stake-"), []byte(owner)...)
		txHashes = append(txHashes, txHash)

		tmp.TxCacher.AddTx(txHash, &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
			Data:    []byte(txData),
		})

		tmp.doStake(t, vmcommon.VMInput{
			CallerAddr:  []byte(owner),
			Arguments:   argsStake,
			CallValue:   nodesData.TotalStake,
			GasProvided: 10,
		})
	}

	blockBody := &block.Body{MiniBlocks: block.MiniBlockSlice{
		{
			TxHashes:        txHashes,
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: core.MetachainShardId,
			Type:            block.SmartContractResultBlock,
		},
	}}
	tmp.TxCoordinator.RequestBlockTransactions(blockBody)
	tmp.createAndCommitBlock(t, header, noTime)

	tmp.currentRound += 1
}

func (tmp *TestMetaProcessor) doStake(t *testing.T, vmInput vmcommon.VMInput) {
	arguments := &vmcommon.ContractCallInput{
		VMInput:       vmInput,
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "stake",
	}
	vmOutput, err := tmp.SystemVM.RunSmartContractCall(arguments)
	require.Nil(t, err)

	err = tmp.processSCOutputAccounts(vmOutput)
	require.Nil(t, err)
}

func createStakingQueueCustomNodes(
	owners map[string]*OwnerStats,
	marshaller marshal.Marshalizer,
	accountsAdapter state.AccountsAdapter,
) [][]byte {
	queue := make([][]byte, 0)

	for owner, ownerStats := range owners {
		stakingcommon.AddKeysToWaitingList(
			accountsAdapter,
			ownerStats.StakingQueueKeys,
			marshaller,
			[]byte(owner),
			[]byte(owner),
		)

		stakingcommon.AddValidatorData(
			accountsAdapter,
			[]byte(owner),
			ownerStats.StakingQueueKeys,
			ownerStats.TotalStake,
			marshaller,
		)

		queue = append(queue, ownerStats.StakingQueueKeys...)
	}

	return queue
}
