package initializers

import (
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"golang.org/x/exp/slices"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string([]byte{byte(i + 65)}))
	}
	return eligibleList
}

// CreateEligibleListFromMap creates a list of eligible nodes from a map of private keys
func CreateEligibleListFromMap(mapKeys map[string]crypto.PrivateKey) []string {
	eligibleList := make([]string, 0, len(mapKeys))
	for key := range mapKeys {
		eligibleList = append(eligibleList, key)
	}
	slices.Sort(eligibleList)
	return eligibleList
}

// InitConsensusStateWithNodesCoordinator creates a consensus state with a nodes coordinator
func InitConsensusStateWithNodesCoordinator(validatorsGroupSelector nodesCoordinator.NodesCoordinator) *spos.ConsensusState {
	return initConsensusStateWithKeysHandlerAndNodesCoordinator(&testscommon.KeysHandlerStub{}, validatorsGroupSelector)
}

// InitConsensusState creates a consensus state
func InitConsensusState() *spos.ConsensusState {
	return InitConsensusStateWithKeysHandler(&testscommon.KeysHandlerStub{})
}

// InitConsensusStateWithArgs creates a consensus state the given arguments
func InitConsensusStateWithArgs(keysHandler consensus.KeysHandler, mapKeys map[string]crypto.PrivateKey) *spos.ConsensusState {
	return initConsensusStateWithKeysHandlerWithGroupSizeWithRealKeys(keysHandler, mapKeys)
}

// InitConsensusStateWithKeysHandler creates a consensus state with a keys handler
func InitConsensusStateWithKeysHandler(keysHandler consensus.KeysHandler) *spos.ConsensusState {
	consensusGroupSize := 9
	return initConsensusStateWithKeysHandlerWithGroupSize(keysHandler, consensusGroupSize)
}

func initConsensusStateWithKeysHandlerAndNodesCoordinator(keysHandler consensus.KeysHandler, validatorsGroupSelector nodesCoordinator.NodesCoordinator) *spos.ConsensusState {
	leader, consensusValidators, _ := validatorsGroupSelector.GetConsensusValidatorsPublicKeys([]byte("randomness"), 0, 0, 0)
	eligibleNodesPubKeys := make(map[string]struct{})
	for _, key := range consensusValidators {
		eligibleNodesPubKeys[key] = struct{}{}
	}
	return createConsensusStateWithNodes(eligibleNodesPubKeys, consensusValidators, leader, keysHandler)
}

// InitConsensusStateWithArgsVerifySignature creates a consensus state with the given arguments for signature verification
func InitConsensusStateWithArgsVerifySignature(keysHandler consensus.KeysHandler, keys []string) *spos.ConsensusState {
	numberOfKeys := len(keys)
	eligibleNodesPubKeys := make(map[string]struct{}, numberOfKeys)
	for _, key := range keys {
		eligibleNodesPubKeys[key] = struct{}{}
	}

	indexLeader := 1
	rcns, _ := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		numberOfKeys,
		keys[indexLeader],
		keysHandler,
	)
	rcns.SetConsensusGroup(keys)
	rcns.ResetRoundState()

	pBFTThreshold := numberOfKeys*2/3 + 1
	pBFTFallbackThreshold := numberOfKeys*1/2 + 1
	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(1, 1)
	rthr.SetThreshold(2, pBFTThreshold)
	rthr.SetFallbackThreshold(1, 1)
	rthr.SetFallbackThreshold(2, pBFTFallbackThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()
	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)
	cns.Data = []byte("X")
	cns.SetRoundIndex(0)

	return cns
}

func initConsensusStateWithKeysHandlerWithGroupSize(keysHandler consensus.KeysHandler, consensusGroupSize int) *spos.ConsensusState {
	eligibleList := createEligibleList(consensusGroupSize)

	eligibleNodesPubKeys := make(map[string]struct{})
	for _, key := range eligibleList {
		eligibleNodesPubKeys[key] = struct{}{}
	}

	return createConsensusStateWithNodes(eligibleNodesPubKeys, eligibleList, eligibleList[0], keysHandler)
}

func initConsensusStateWithKeysHandlerWithGroupSizeWithRealKeys(keysHandler consensus.KeysHandler, mapKeys map[string]crypto.PrivateKey) *spos.ConsensusState {
	eligibleList := CreateEligibleListFromMap(mapKeys)

	eligibleNodesPubKeys := make(map[string]struct{}, len(eligibleList))
	for _, key := range eligibleList {
		eligibleNodesPubKeys[key] = struct{}{}
	}

	return createConsensusStateWithNodes(eligibleNodesPubKeys, eligibleList, eligibleList[0], keysHandler)
}

func createConsensusStateWithNodes(eligibleNodesPubKeys map[string]struct{}, consensusValidators []string, leader string, keysHandler consensus.KeysHandler) *spos.ConsensusState {
	consensusGroupSize := len(consensusValidators)
	rcns, _ := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		consensusGroupSize,
		consensusValidators[1],
		keysHandler,
	)

	rcns.SetConsensusGroup(consensusValidators)
	rcns.SetLeader(leader)
	rcns.ResetRoundState()

	pBFTThreshold := consensusGroupSize*2/3 + 1
	pBFTFallbackThreshold := consensusGroupSize*1/2 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(1, 1)
	rthr.SetThreshold(2, pBFTThreshold)
	rthr.SetFallbackThreshold(1, 1)
	rthr.SetFallbackThreshold(2, pBFTFallbackThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")
	cns.SetRoundIndex(0)
	return cns
}
