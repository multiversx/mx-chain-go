package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

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
	_, err := tmp.MetaBlockProcessor.CreateNewHeader(tmp.currentRound, tmp.currentRound)
	require.Nil(t, err)

	epoch := tmp.EpochStartTrigger.Epoch()
	printNewHeaderRoundEpoch(tmp.currentRound, epoch)

	currentHeader, currentHash := tmp.getCurrentHeaderInfo()
	header := createMetaBlockToCommit(
		epoch,
		tmp.currentRound,
		currentHash,
		currentHeader.GetRandSeed(),
		tmp.NodesCoordinator.ConsensusGroupSize(core.MetachainShardId),
	)
	tmp.BlockChainHook.SetCurrentHeader(header)

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	blockBody := &block.Body{MiniBlocks: make([]*block.MiniBlock, 0)}

	for owner, nodesData := range nodes {
		numBLSKeys := int64(len(nodesData.BLSKeys))
		numOfNodesToStake := big.NewInt(numBLSKeys).Bytes()
		numOfNodesToStakeHex := hex.EncodeToString(numOfNodesToStake)
		_ = numOfNodesToStakeHex
		for _, blsKey := range nodesData.BLSKeys {
			signature := append([]byte("signature-"), blsKey...)
			txData := hex.EncodeToString([]byte("stake")) + "@" +
				hex.EncodeToString(big.NewInt(1).Bytes()) + "@" +
				hex.EncodeToString(blsKey) + "@" +
				hex.EncodeToString(signature)

			mbHeaderHash := []byte(fmt.Sprintf("mbHash-stake-blsKey=%s-owner=%s", blsKey, owner))
			shardMiniBlockHeader := block.MiniBlockHeader{
				Hash:            mbHeaderHash,
				ReceiverShardID: 0,
				SenderShardID:   core.MetachainShardId,
				TxCount:         1,
			}
			shardMiniBlockHeaders = append(header.MiniBlockHeaders, shardMiniBlockHeader)
			shardData := block.ShardData{
				Nonce:                 tmp.currentRound,
				ShardID:               0,
				HeaderHash:            []byte("hdr_hashStake"),
				TxCount:               1,
				ShardMiniBlockHeaders: shardMiniBlockHeaders,
			}
			header.ShardInfo = append(header.ShardInfo, shardData)
			tmp.TxCacher.AddTx(mbHeaderHash, &smartContractResult.SmartContractResult{
				RcvAddr: vm.StakingSCAddress,
				Data:    []byte(txData),
			})

			blockBody.MiniBlocks = append(blockBody.MiniBlocks, &block.MiniBlock{
				TxHashes:        [][]byte{mbHeaderHash},
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: core.MetachainShardId,
				Type:            block.SmartContractResultBlock,
			},
			)

			arguments := &vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr:  []byte(owner),
					Arguments:   [][]byte{big.NewInt(1).Bytes(), blsKey, signature},
					CallValue:   big.NewInt(nodesData.TotalStake.Int64()).Div(nodesData.TotalStake, big.NewInt(numBLSKeys)),
					GasProvided: 10,
				},
				RecipientAddr: vm.ValidatorSCAddress,
				Function:      "stake",
			}
			vmOutput, _ := tmp.SystemVM.RunSmartContractCall(arguments)

			_, _ = tmp.processSCOutputAccounts(vmOutput)
		}

	}
	tmp.TxCoordinator.RequestBlockTransactions(blockBody)

	haveTime := func() bool { return false }
	newHeader, newBlockBody, err := tmp.MetaBlockProcessor.CreateBlock(header, haveTime)
	require.Nil(t, err)

	err = tmp.MetaBlockProcessor.CommitBlock(newHeader, newBlockBody)
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 50)
	tmp.updateNodesConfig(epoch)
	displayConfig(tmp.NodesConfig)

	tmp.currentRound += 1
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
