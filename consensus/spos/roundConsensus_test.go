package spos_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func initRoundConsensus() *spos.RoundConsensus {
	return initRoundConsensusWithKeysHandler(&testscommon.KeysHandlerStub{})
}

func initRoundConsensusWithKeysHandler(keysHandler consensus.KeysHandler) *spos.RoundConsensus {
	pubKeys := []string{"1", "2", "3"}
	eligibleNodes := make(map[string]struct{})

	for i := range pubKeys {
		eligibleNodes[pubKeys[i]] = struct{}{}
	}

	rcns, _ := spos.NewRoundConsensus(
		eligibleNodes,
		len(eligibleNodes),
		"2",
		keysHandler,
	)

	rcns.SetConsensusGroup(pubKeys)

	rcns.ResetRoundState()

	return spos.NewRoundConsensusWrapper(rcns)
}

func TestRoundConsensus_NewRoundConsensusShouldWork(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.NotNil(t, rcns)
	assert.Equal(t, 3, len(rcns.ConsensusGroup()))
	assert.Equal(t, "3", rcns.ConsensusGroup()[2])
	assert.Equal(t, "2", rcns.SelfPubKey())
}

func TestRoundConsensus_ConsensusGroupIndexFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}
	eligibleNodes := make(map[string]struct{})

	for i := range pubKeys {
		eligibleNodes[pubKeys[i]] = struct{}{}
	}

	rcns, _ := spos.NewRoundConsensus(eligibleNodes, 3, "key3", &testscommon.KeysHandlerStub{})
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key3")

	assert.Equal(t, 2, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_ConsensusGroupIndexNotFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}
	eligibleNodes := make(map[string]struct{})

	for i := range pubKeys {
		eligibleNodes[pubKeys[i]] = struct{}{}
	}

	rcns, _ := spos.NewRoundConsensus(eligibleNodes, 3, "key4", &testscommon.KeysHandlerStub{})
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key4")

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupInConsesus(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}
	eligibleNodes := make(map[string]struct{})

	for i := range pubKeys {
		eligibleNodes[pubKeys[i]] = struct{}{}
	}

	rcns, _ := spos.NewRoundConsensus(eligibleNodes, 3, "key2", &testscommon.KeysHandlerStub{})
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.SelfConsensusGroupIndex()

	assert.Equal(t, 1, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupNotFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}
	eligibleNodes := make(map[string]struct{})

	for i := range pubKeys {
		eligibleNodes[pubKeys[i]] = struct{}{}
	}

	rcns, _ := spos.NewRoundConsensus(eligibleNodes, 3, "key4", &testscommon.KeysHandlerStub{})
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.SelfConsensusGroupIndex()

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestRoundConsensus_SetEligibleListShouldChangeTheEligibleList(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()
	eligibleList := []string{"4", "5", "6"}

	eligibleNodesKeys := make(map[string]struct{})
	for _, key := range eligibleList {
		eligibleNodesKeys[key] = struct{}{}
	}

	rcns.SetEligibleList(eligibleNodesKeys)

	_, ok := rcns.EligibleList()["4"]
	assert.True(t, ok)
	_, ok = rcns.EligibleList()["5"]
	assert.True(t, ok)
	_, ok = rcns.EligibleList()["6"]
	assert.True(t, ok)
}

func TestRoundConsensus_SetConsensusGroupShouldChangeTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetConsensusGroup([]string{"4", "5", "6"})

	assert.Equal(t, "4", rcns.ConsensusGroup()[0])
	assert.Equal(t, "5", rcns.ConsensusGroup()[1])
	assert.Equal(t, "6", rcns.ConsensusGroup()[2])
}

func TestRoundConsensus_SetConsensusGroupSizeShouldChangeTheConsensusGroupSize(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, len(rcns.ConsensusGroup()), rcns.ConsensusGroupSize())
	rcns.SetConsensusGroupSize(99999)
	assert.Equal(t, 99999, rcns.ConsensusGroupSize())
}

func TestRoundConsensus_SetSelfPubKeyShouldChangeTheSelfPubKey(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetSelfPubKey("X")
	assert.Equal(t, "X", rcns.SelfPubKey())
}

func TestRoundConsensus_GetJobDoneShouldReturnsFalseWhenValidatorIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	_ = rcns.SetJobDone("3", bls.SrBlock, true)
	rcns.SetConsensusGroup([]string{"1", "2"})
	isJobDone, _ := rcns.JobDone("3", bls.SrBlock)
	assert.False(t, isJobDone)
}

func TestRoundConsensus_SetJobDoneShouldNotBeSetWhenValidatorIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	_ = rcns.SetJobDone("4", bls.SrBlock, true)
	isJobDone, _ := rcns.JobDone("4", bls.SrBlock)
	assert.False(t, isJobDone)
}

func TestRoundConsensus_GetSelfJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		if rcns.ConsensusGroup()[i] == rcns.SelfPubKey() {
			continue
		}

		_ = rcns.SetJobDone(rcns.ConsensusGroup()[i], bls.SrBlock, true)
	}

	jobDone, _ := rcns.SelfJobDone(bls.SrBlock)
	assert.False(t, jobDone)
}

func TestRoundConsensus_GetSelfJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	_ = rcns.SetJobDone("2", bls.SrBlock, true)

	jobDone, _ := rcns.SelfJobDone(bls.SrBlock)
	assert.True(t, jobDone)
}

func TestRoundConsensus_SetSelfJobDoneShouldWork(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	_ = rcns.SetJobDone(rcns.SelfPubKey(), bls.SrBlock, true)

	jobDone, _ := rcns.JobDone("2", bls.SrBlock)
	assert.True(t, jobDone)
}

func TestRoundConsensus_IsNodeInConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, false, rcns.IsNodeInConsensusGroup("4"))
	assert.Equal(t, true, rcns.IsNodeInConsensusGroup(rcns.SelfPubKey()))
}

func TestRoundConsensus_IsNodeInEligibleList(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, false, rcns.IsNodeInEligibleList("4"))
	assert.Equal(t, true, rcns.IsNodeInEligibleList(rcns.SelfPubKey()))
}

func TestRoundConsensus_ComputeSize(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	_ = rcns.SetJobDone("1", bls.SrBlock, true)
	assert.Equal(t, 1, rcns.ComputeSize(bls.SrBlock))
}

func TestRoundConsensus_ResetValidationMap(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	_ = rcns.SetJobDone("1", bls.SrBlock, true)
	jobDone, _ := rcns.JobDone("1", bls.SrBlock)
	assert.Equal(t, true, jobDone)

	rcns.ConsensusGroup()[1] = "X"

	rcns.ResetRoundState()

	jobDone, err := rcns.JobDone("1", bls.SrBlock)
	assert.Equal(t, false, jobDone)
	assert.Nil(t, err)
}

func TestRoundConsensus_IsMultiKeyInConsensusGroup(t *testing.T) {
	t.Parallel()

	keysHandler := &testscommon.KeysHandlerStub{}
	roundConsensus := initRoundConsensusWithKeysHandler(keysHandler)
	t.Run("no consensus key is managed by current node should return false", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return false
		}
		assert.False(t, roundConsensus.IsMultiKeyInConsensusGroup())
	})
	t.Run("consensus key is managed by current node should return true", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("2"), pkBytes)
		}
		assert.True(t, roundConsensus.IsMultiKeyInConsensusGroup())
	})
}

func TestRoundConsensus_IsKeyManagedByCurrentNode(t *testing.T) {
	t.Parallel()

	managedPkBytes := []byte("managed pk bytes")
	wasCalled := false
	keysHandler := &testscommon.KeysHandlerStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			assert.Equal(t, managedPkBytes, pkBytes)
			wasCalled = true
			return true
		},
	}
	roundConsensus := initRoundConsensusWithKeysHandler(keysHandler)
	assert.True(t, roundConsensus.IsKeyManagedByCurrentNode(managedPkBytes))
	assert.True(t, wasCalled)
}

func TestRoundConsensus_IncrementRoundsWithoutReceivedMessages(t *testing.T) {
	t.Parallel()

	managedPkBytes := []byte("managed pk bytes")
	wasCalled := false
	keysHandler := &testscommon.KeysHandlerStub{
		IncrementRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte) {
			assert.Equal(t, managedPkBytes, pkBytes)
			wasCalled = true
		},
	}
	roundConsensus := initRoundConsensusWithKeysHandler(keysHandler)
	roundConsensus.IncrementRoundsWithoutReceivedMessages(managedPkBytes)
	assert.True(t, wasCalled)
}
