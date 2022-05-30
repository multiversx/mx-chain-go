package staking

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

// OwnerStats -
type OwnerStats struct {
	EligibleBlsKeys  map[uint32][][]byte
	WaitingBlsKeys   map[uint32][][]byte
	StakingQueueKeys [][]byte
	TotalStake       *big.Int
}

// InitialNodesConfig -
type InitialNodesConfig struct {
	Owners                        map[string]*OwnerStats
	MaxNodesChangeConfig          []config.MaxNodesChangeConfig
	NumOfShards                   uint32
	MinNumberOfEligibleShardNodes uint32
	MinNumberOfEligibleMetaNodes  uint32
	ShardConsensusGroupSize       int
	MetaConsensusGroupSize        int
}

// NewTestMetaProcessorWithCustomNodes -
func NewTestMetaProcessorWithCustomNodes(config *InitialNodesConfig) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(config.NumOfShards)

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

// NodesRegisterData -
type NodesRegisterData struct {
	BLSKeys    [][]byte
	TotalStake *big.Int
}

// ProcessStake will create a block containing mini blocks with staking txs using provided nodes.
// Block will be committed + call to validator system sc will be made to stake all nodes
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
	_, err := tmp.AccountsAdapter.Commit()
	require.Nil(t, err)

	miniBlocks := block.MiniBlockSlice{
		{
			TxHashes:        txHashes,
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: core.MetachainShardId,
			Type:            block.SmartContractResultBlock,
		},
	}
	tmp.TxCoordinator.AddTxsFromMiniBlocks(miniBlocks)
	tmp.createAndCommitBlock(t, header, noTime)

	tmp.currentRound += 1
}

// ProcessUnStake will create a block containing mini blocks with unStaking txs using provided nodes.
// Block will be committed + call to validator system sc will be made to unStake all nodes
func (tmp *TestMetaProcessor) ProcessUnStake(t *testing.T, nodes map[string]*NodesRegisterData) {
	header := tmp.createNewHeader(t, tmp.currentRound)
	tmp.BlockChainHook.SetCurrentHeader(header)

	txHashes := make([][]byte, 0)
	for owner, nodesData := range nodes {
		numBLSKeys := int64(len(nodesData.BLSKeys))
		numBLSKeysBytes := big.NewInt(numBLSKeys).Bytes()

		txData := hex.EncodeToString([]byte("unStake")) + "@" + hex.EncodeToString(numBLSKeysBytes)
		argsUnStake := make([][]byte, 0)

		for _, blsKey := range nodesData.BLSKeys {
			argsUnStake = append(argsUnStake, blsKey)
			txData += "@" + hex.EncodeToString(blsKey) + "@"
		}

		txHash := append([]byte("txHash-unStake-"), []byte(owner)...)
		txHashes = append(txHashes, txHash)

		tmp.TxCacher.AddTx(txHash, &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
			Data:    []byte(txData),
		})

		txsData := tmp.doUnStake(t, vmcommon.VMInput{
			CallerAddr:  []byte(owner),
			Arguments:   argsUnStake,
			CallValue:   big.NewInt(0),
			GasProvided: 10,
		})

		for i, tData := range txsData {
			txHash = []byte("rrrr" + strconv.Itoa(i))
			txHashes = append(txHashes, txHash)

			tmp.TxCacher.AddTx(txHash, &smartContractResult.SmartContractResult{
				RcvAddr: vm.StakingSCAddress,
				Data:    []byte(tData),
			})

		}
	}
	_, err := tmp.AccountsAdapter.Commit()
	require.Nil(t, err)

	miniBlocks := block.MiniBlockSlice{
		{
			TxHashes:        txHashes,
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: core.MetachainShardId,
			Type:            block.SmartContractResultBlock,
		},
	}
	tmp.TxCoordinator.AddTxsFromMiniBlocks(miniBlocks)
	tmp.createAndCommitBlock(t, header, noTime)

	tmp.currentRound += 1
}

//TODO:
// - Do the same for unStake/unJail
func (tmp *TestMetaProcessor) doStake(t *testing.T, vmInput vmcommon.VMInput) {
	arguments := &vmcommon.ContractCallInput{
		VMInput:       vmInput,
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "stake",
	}
	vmOutput, err := tmp.SystemVM.RunSmartContractCall(arguments)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	err = integrationTests.ProcessSCOutputAccounts(vmOutput, tmp.AccountsAdapter)
	require.Nil(t, err)
}

func (tmp *TestMetaProcessor) doUnStake(t *testing.T, vmInput vmcommon.VMInput) []string {
	arguments := &vmcommon.ContractCallInput{
		VMInput:       vmInput,
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "unStake",
	}
	vmOutput, err := tmp.SystemVM.RunSmartContractCall(arguments)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	txsData, err := ProcessSCOutputAccounts(vmOutput, tmp.AccountsAdapter)
	require.Nil(t, err)
	return txsData
}

func ProcessSCOutputAccounts(vmOutput *vmcommon.VMOutput, accountsDB state.AccountsAdapter) ([]string, error) {
	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	data := make([]string, 0)
	for _, outAcc := range outputAccounts {
		acc := stakingcommon.LoadUserAccount(accountsDB, outAcc.Address)

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
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

			err = accountsDB.SaveAccount(acc)
			if err != nil {
				return nil, err
			}

			if bytes.Equal(outAcc.Address, vm.StakingSCAddress) {
				parser := smartContract.NewArgumentParser()
				data2 := parser.CreateDataFromStorageUpdate(storageUpdates)
				data = append(data, data2)

			}

		}
	}

	return data, nil
}
