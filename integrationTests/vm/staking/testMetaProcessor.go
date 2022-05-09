package staking

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/display"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

const (
	stakingV4InitEpoch                       = 1
	stakingV4EnableEpoch                     = 2
	stakingV4DistributeAuctionToWaitingEpoch = 3
	addressLength                            = 15
	nodePrice                                = 1000
)

type nodesConfig struct {
	eligible    map[uint32][][]byte
	waiting     map[uint32][][]byte
	leaving     map[uint32][][]byte
	shuffledOut map[uint32][][]byte
	queue       [][]byte
	auction     [][]byte
}

// TestMetaProcessor -
type TestMetaProcessor struct {
	MetaBlockProcessor  process.BlockProcessor
	NodesCoordinator    nodesCoordinator.NodesCoordinator
	ValidatorStatistics process.ValidatorStatisticsProcessor
	EpochStartTrigger   integrationTests.TestEpochStartTrigger
	BlockChainHandler   data.ChainHandler
	NodesConfig         nodesConfig
	AccountsAdapter     state.AccountsAdapter
	Marshaller          marshal.Marshalizer
	TxCacher            dataRetriever.TransactionCacher
	TxCoordinator       process.TransactionCoordinator
	SystemVM            vmcommon.VMExecutionHandler
	StateComponents     factory.StateComponentsHolder
	BlockChainHook      process.BlockChainHookHandler

	currentRound uint64
}

// NewTestMetaProcessor -
func NewTestMetaProcessor(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	numOfNodesToShufflePerShard uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	numOfNodesInStakingQueue uint32,
) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(numOfShards)

	maxNodesConfig := createMaxNodesConfig(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		numOfNodesToShufflePerShard,
	)

	queue := createStakingQueue(
		numOfNodesInStakingQueue,
		maxNodesConfig[0].MaxNumNodes,
		coreComponents.InternalMarshalizer(),
		stateComponents.AccountsAdapter(),
	)

	eligibleMap, waitingMap := createGenesisNodes(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		coreComponents.InternalMarshalizer(),
		stateComponents,
	)

	nc := createNodesCoordinator(
		eligibleMap,
		waitingMap,
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		coreComponents,
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		bootstrapComponents.NodesCoordinatorRegistryFactory(),
		maxNodesConfig,
	)

	return newTestMetaProcessor(
		coreComponents,
		dataComponents,
		bootstrapComponents,
		statusComponents,
		stateComponents,
		nc,
		maxNodesConfig,
		queue,
	)
}

func createMaxNodesConfig(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	numOfNodesToShufflePerShard uint32,
) []config.MaxNodesChangeConfig {
	totalEligible := numOfMetaNodes + numOfShards*numOfEligibleNodesPerShard
	totalWaiting := (numOfShards + 1) * numOfWaitingNodesPerShard
	totalNodes := totalEligible + totalWaiting

	maxNodesConfig := make([]config.MaxNodesChangeConfig, 0)
	maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            totalNodes,
		NodesToShufflePerShard: numOfNodesToShufflePerShard,
	},
	)

	maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{
		EpochEnable:            stakingV4DistributeAuctionToWaitingEpoch,
		MaxNumNodes:            totalNodes - numOfNodesToShufflePerShard*(numOfShards+1),
		NodesToShufflePerShard: numOfNodesToShufflePerShard,
	},
	)

	return maxNodesConfig
}

// Process -
func (tmp *TestMetaProcessor) Process(t *testing.T, numOfRounds uint64) {
	for r := tmp.currentRound; r < tmp.currentRound+numOfRounds; r++ {
		_, err := tmp.MetaBlockProcessor.CreateNewHeader(r, r)
		require.Nil(t, err)

		epoch := tmp.EpochStartTrigger.Epoch()
		printNewHeaderRoundEpoch(r, epoch)

		currentHeader, currentHash := tmp.getCurrentHeaderInfo()
		header := createMetaBlockToCommit(
			epoch,
			r,
			currentHash,
			currentHeader.GetRandSeed(),
			tmp.NodesCoordinator.ConsensusGroupSize(core.MetachainShardId),
		)

		haveTime := func() bool { return true }

		if r == 17 && numOfRounds == 25 {
			oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
			pubKey := hex.EncodeToString([]byte("000address-3198"))
			txData := hex.EncodeToString([]byte("stake")) + "@" + oneEncoded + "@" + pubKey + "@" + hex.EncodeToString([]byte("signature"))

			shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
			shardMiniBlockHeader := block.MiniBlockHeader{
				Hash:            []byte("hashStake"),
				ReceiverShardID: 0,
				SenderShardID:   core.MetachainShardId,
				TxCount:         1,
			}
			shardMiniBlockHeaders = append(header.MiniBlockHeaders, shardMiniBlockHeader)
			shardData := block.ShardData{
				Nonce:                 r,
				ShardID:               0,
				HeaderHash:            []byte("hdr_hashStake"),
				TxCount:               1,
				ShardMiniBlockHeaders: shardMiniBlockHeaders,
				DeveloperFees:         big.NewInt(0),
				AccumulatedFees:       big.NewInt(0),
			}
			header.ShardInfo = append(header.ShardInfo, shardData)
			tmp.TxCacher.AddTx(shardMiniBlockHeader.Hash, &smartContractResult.SmartContractResult{
				RcvAddr: vm.StakingSCAddress,
				Data:    []byte(txData),
			})

			haveTime = func() bool { return false }

			blockBody := &block.Body{
				MiniBlocks: []*block.MiniBlock{
					{
						TxHashes:        [][]byte{shardMiniBlockHeader.Hash},
						SenderShardID:   core.MetachainShardId,
						ReceiverShardID: core.MetachainShardId,
						Type:            block.SmartContractResultBlock,
					},
				},
			}

			tmp.TxCoordinator.RequestBlockTransactions(blockBody)

			tmp.BlockChainHook.SetCurrentHeader(header)

			arguments := &vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr: vm.EndOfEpochAddress,
					CallValue:  big.NewInt(0),
				},
				RecipientAddr: vm.StakingSCAddress,
			}
			arguments.Function = "stake"
			arguments.CallerAddr = []byte("000address-3198")
			arguments.RecipientAddr = vm.ValidatorSCAddress
			arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("000address-3198"), []byte("signature")}
			arguments.CallValue = big.NewInt(2000)
			arguments.GasProvided = 10

			vmOutput, _ := tmp.SystemVM.RunSmartContractCall(arguments)

			stakedData, _ := tmp.processSCOutputAccounts(vmOutput)
			//stakingSC := stakingcommon.LoadUserAccount(tmp.AccountsAdapter, vm.StakingSCAddress)
			//stakedDataBuffer, _ := stakingSC.DataTrieTracker().RetrieveValue([]byte("000address-3198"))
			//
			//_ = stakingSC.DataTrieTracker().SaveKeyValue([]byte("000address-3198"), stakedData)
			//
			//tmp.AccountsAdapter.SaveAccount(stakingSC)

			//var peerAcc state.PeerAccountHandler
			//
			//peerAcc, _ = state.NewPeerAccount([]byte("000address-3198"))
			//
			//tmp.StateComponents.PeerAccounts().SaveAccount(peerAcc)
			//tmp.AccountsAdapter.SaveAccount(peerAcc)
			//
			//tmp.AccountsAdapter.Commit()
			//tmp.StateComponents.PeerAccounts().Commit()
			//
			//loadedAcc, _ := tmp.StateComponents.PeerAccounts().LoadAccount([]byte("000address-3198"))
			//
			//loadedAccCasted, castOK := loadedAcc.(state.PeerAccountHandler)
			//if castOK {
			//
			//}

			/*
				stakingcommon.AddValidatorData(
					tmp.AccountsAdapter,
					[]byte("000address-3198"),
					[][]byte{[]byte("000address-3198")},
					big.NewInt(1000),
					tmp.Marshaller,
				)

			*/

			tmp.AccountsAdapter.Commit()
			tmp.StateComponents.PeerAccounts().Commit()

			//stakedDataBuffer, _ = stakingSC.DataTrieTracker().RetrieveValue([]byte("000address-3198"))
			//_ = stakedDataBuffer
			_ = vmOutput
			_ = stakedData
			//_ = loadedAcc
			//_ = loadedAccCasted
		}

		newHeader, blockBody, err := tmp.MetaBlockProcessor.CreateBlock(header, haveTime)
		require.Nil(t, err)

		err = tmp.MetaBlockProcessor.CommitBlock(newHeader, blockBody)
		require.Nil(t, err)

		time.Sleep(time.Millisecond * 50)
		tmp.updateNodesConfig(epoch)
		displayConfig(tmp.NodesConfig)
	}

	tmp.currentRound += numOfRounds
}

func printNewHeaderRoundEpoch(round uint64, epoch uint32) {
	headline := display.Headline(
		fmt.Sprintf("Commiting header in epoch %v round %v", epoch, round),
		"",
		delimiter,
	)
	fmt.Println(headline)
}

func (tmp *TestMetaProcessor) getCurrentHeaderInfo() (data.HeaderHandler, []byte) {
	currentHeader := tmp.BlockChainHandler.GetCurrentBlockHeader()
	currentHash := tmp.BlockChainHandler.GetCurrentBlockHeaderHash()
	if currentHeader == nil {
		currentHeader = tmp.BlockChainHandler.GetGenesisHeader()
		currentHash = tmp.BlockChainHandler.GetGenesisHeaderHash()
	}

	return currentHeader, currentHash
}

func createMetaBlockToCommit(
	epoch uint32,
	round uint64,
	prevHash []byte,
	prevRandSeed []byte,
	consensusSize int,
) *block.MetaBlock {
	roundStr := strconv.Itoa(int(round))
	hdr := block.MetaBlock{
		Epoch:                  epoch,
		Nonce:                  round,
		Round:                  round,
		PrevHash:               prevHash,
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte(strings.Repeat("f", consensusSize)),
		RootHash:               []byte("roothash" + roundStr),
		ShardInfo:              make([]block.ShardData, 0),
		TxCount:                1,
		PrevRandSeed:           prevRandSeed,
		RandSeed:               []byte("randseed" + roundStr),
		AccumulatedFeesInEpoch: big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
	}

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	shardMiniBlockHeader := block.MiniBlockHeader{
		Hash:            []byte("mb_hash" + roundStr),
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		Nonce:                 round,
		ShardID:               0,
		HeaderHash:            []byte("hdr_hash" + roundStr),
		TxCount:               1,
		ShardMiniBlockHeaders: shardMiniBlockHeaders,
		DeveloperFees:         big.NewInt(0),
		AccumulatedFees:       big.NewInt(0),
	}
	hdr.ShardInfo = append(hdr.ShardInfo, shardData)

	return &hdr
}

func (tmp *TestMetaProcessor) updateNodesConfig(epoch uint32) {
	eligible, _ := tmp.NodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	waiting, _ := tmp.NodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	leaving, _ := tmp.NodesCoordinator.GetAllLeavingValidatorsPublicKeys(epoch)
	shuffledOut, _ := tmp.NodesCoordinator.GetAllShuffledOutValidatorsPublicKeys(epoch)

	rootHash, _ := tmp.ValidatorStatistics.RootHash()
	validatorsInfoMap, _ := tmp.ValidatorStatistics.GetValidatorInfoForRootHash(rootHash)

	auction := make([][]byte, 0)
	for _, validator := range validatorsInfoMap.GetAllValidatorsInfo() {
		if validator.GetList() == string(common.AuctionList) {
			auction = append(auction, validator.GetPublicKey())
		}
	}

	tmp.NodesConfig.eligible = eligible
	tmp.NodesConfig.waiting = waiting
	tmp.NodesConfig.shuffledOut = shuffledOut
	tmp.NodesConfig.leaving = leaving
	tmp.NodesConfig.auction = auction
	tmp.NodesConfig.queue = tmp.getWaitingListKeys()
}

func generateAddresses(startIdx, n uint32) [][]byte {
	ret := make([][]byte, 0, n)

	for i := startIdx; i < n+startIdx; i++ {
		ret = append(ret, generateAddress(i))
	}

	return ret
}

func generateAddress(identifier uint32) []byte {
	uniqueIdentifier := fmt.Sprintf("address-%d", identifier)
	return []byte(strings.Repeat("0", addressLength-len(uniqueIdentifier)) + uniqueIdentifier)
}

func (tmp *TestMetaProcessor) processSCOutputAccounts(vmOutput *vmcommon.VMOutput) ([]byte, error) {
	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		if bytes.Equal(outAcc.Address, vm.StakingSCAddress) {
			fmt.Println("DSADA")
		}
		if bytes.Equal(outAcc.Address, vm.ValidatorSCAddress) {
			fmt.Println("VAAAAAAAAAAAAAAAAAAAAALLLLLLLLLLLLLl")
		}

		acc := stakingcommon.LoadUserAccount(tmp.AccountsAdapter, outAcc.Address)

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
			if bytes.Equal(storeUpdate.Offset, []byte("000address-3198")) {
				fmt.Println("DASDSA")
				//return storeUpdate.Data, nil
			}
			err := acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				return nil, err
			}

			if outAcc.BalanceDelta != nil && outAcc.BalanceDelta.Cmp(big.NewInt(0)) != 0 {
				err = acc.AddToBalance(outAcc.BalanceDelta)
				if err != nil {
					return nil, err
				}
			}

			err = tmp.AccountsAdapter.SaveAccount(acc)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
