package staking

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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

	bootstrapStorer, _ := dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	nc := createNodesCoordinator(
		eligibleMap,
		waitingMap,
		config.MinNumberOfEligibleMetaNodes,
		config.NumOfShards,
		config.MinNumberOfEligibleShardNodes,
		config.ShardConsensusGroupSize,
		config.MetaConsensusGroupSize,
		coreComponents,
		bootstrapStorer,
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
	for owner, registerData := range nodes {
		scrs := tmp.doStake(t, []byte(owner), registerData)
		txHashes = append(txHashes, tmp.addTxsToCacher(scrs)...)
	}

	tmp.commitBlockTxs(t, txHashes, header)
}

func (tmp *TestMetaProcessor) doStake(
	t *testing.T,
	owner []byte,
	registerData *NodesRegisterData,
) map[string]*smartContractResult.SmartContractResult {
	arguments := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  owner,
			Arguments:   createStakeArgs(registerData.BLSKeys),
			CallValue:   registerData.TotalStake,
			GasProvided: 10,
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "stake",
	}

	return tmp.runSC(t, arguments)
}

func createStakeArgs(blsKeys [][]byte) [][]byte {
	numBLSKeys := int64(len(blsKeys))
	numBLSKeysBytes := big.NewInt(numBLSKeys).Bytes()
	argsStake := [][]byte{numBLSKeysBytes}

	for _, blsKey := range blsKeys {
		signature := append([]byte("signature-"), blsKey...)
		argsStake = append(argsStake, blsKey, signature)
	}

	return argsStake
}

// ProcessUnStake will create a block containing mini blocks with unStaking txs using provided nodes.
// Block will be committed + call to validator system sc will be made to unStake all nodes
func (tmp *TestMetaProcessor) ProcessUnStake(t *testing.T, nodes map[string][][]byte) {
	header := tmp.createNewHeader(t, tmp.currentRound)
	tmp.BlockChainHook.SetCurrentHeader(header)

	txHashes := make([][]byte, 0)
	for owner, blsKeys := range nodes {
		scrs := tmp.doUnStake(t, []byte(owner), blsKeys)
		txHashes = append(txHashes, tmp.addTxsToCacher(scrs)...)
	}

	tmp.commitBlockTxs(t, txHashes, header)
}

func (tmp *TestMetaProcessor) doUnStake(
	t *testing.T,
	owner []byte,
	blsKeys [][]byte,
) map[string]*smartContractResult.SmartContractResult {
	arguments := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  owner,
			Arguments:   blsKeys,
			CallValue:   big.NewInt(0),
			GasProvided: 10,
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "unStake",
	}

	return tmp.runSC(t, arguments)
}

// ProcessJail will create a block containing mini blocks with jail txs using provided nodes.
// Block will be committed + call to validator system sc will be made to jail all nodes
func (tmp *TestMetaProcessor) ProcessJail(t *testing.T, blsKeys [][]byte) {
	header := tmp.createNewHeader(t, tmp.currentRound)
	tmp.BlockChainHook.SetCurrentHeader(header)

	scrs := tmp.doJail(t, blsKeys)
	txHashes := tmp.addTxsToCacher(scrs)
	tmp.commitBlockTxs(t, txHashes, header)
}

func (tmp *TestMetaProcessor) doJail(
	t *testing.T,
	blsKeys [][]byte,
) map[string]*smartContractResult.SmartContractResult {
	arguments := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.JailingAddress,
			Arguments:   blsKeys,
			CallValue:   big.NewInt(0),
			GasProvided: 10,
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "jail",
	}

	return tmp.runSC(t, arguments)
}

// ProcessUnJail will create a block containing mini blocks with unJail txs using provided nodes.
// Block will be committed + call to validator system sc will be made to unJail all nodes
func (tmp *TestMetaProcessor) ProcessUnJail(t *testing.T, blsKeys [][]byte) {
	header := tmp.createNewHeader(t, tmp.currentRound)
	tmp.BlockChainHook.SetCurrentHeader(header)

	txHashes := make([][]byte, 0)
	for _, blsKey := range blsKeys {
		scrs := tmp.doUnJail(t, blsKey)
		txHashes = append(txHashes, tmp.addTxsToCacher(scrs)...)
	}

	tmp.commitBlockTxs(t, txHashes, header)
}

func (tmp *TestMetaProcessor) doUnJail(
	t *testing.T,
	blsKey []byte,
) map[string]*smartContractResult.SmartContractResult {
	arguments := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.ValidatorSCAddress,
			Arguments:   [][]byte{blsKey},
			CallValue:   big.NewInt(0),
			GasProvided: 10,
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "unJail",
	}

	return tmp.runSC(t, arguments)
}

func (tmp *TestMetaProcessor) addTxsToCacher(scrs map[string]*smartContractResult.SmartContractResult) [][]byte {
	txHashes := make([][]byte, 0)
	for scrHash, scr := range scrs {
		txHashes = append(txHashes, []byte(scrHash))
		tmp.TxCacher.AddTx([]byte(scrHash), scr)
	}

	return txHashes
}

func (tmp *TestMetaProcessor) commitBlockTxs(t *testing.T, txHashes [][]byte, header data.HeaderHandler) {
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

func (tmp *TestMetaProcessor) runSC(t *testing.T, arguments *vmcommon.ContractCallInput) map[string]*smartContractResult.SmartContractResult {
	vmOutput, err := tmp.SystemVM.RunSmartContractCall(arguments)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	err = integrationTests.ProcessSCOutputAccounts(vmOutput, tmp.AccountsAdapter)
	require.Nil(t, err)

	return createSCRsFromStakingSCOutput(vmOutput, tmp.Marshaller)
}

func createSCRsFromStakingSCOutput(
	vmOutput *vmcommon.VMOutput,
	marshaller marshal.Marshalizer,
) map[string]*smartContractResult.SmartContractResult {
	allSCR := make(map[string]*smartContractResult.SmartContractResult)
	parser := smartContract.NewArgumentParser()
	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		storageUpdates := process.GetSortedStorageUpdates(outAcc)

		if bytes.Equal(outAcc.Address, vm.StakingSCAddress) {
			scrData := parser.CreateDataFromStorageUpdate(storageUpdates)
			scr := &smartContractResult.SmartContractResult{
				RcvAddr: vm.StakingSCAddress,
				Data:    []byte(scrData),
			}
			scrBytes, _ := marshaller.Marshal(scr)
			scrHash := hex.EncodeToString(scrBytes)

			allSCR[scrHash] = scr
		}
	}

	return allSCR
}
