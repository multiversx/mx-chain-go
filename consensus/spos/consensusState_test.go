package spos_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func internalInitConsensusState() *spos.ConsensusState {
	return internalInitConsensusStateWithKeysHandler(&testscommon.KeysHandlerStub{})
}

func internalInitConsensusStateWithKeysHandler(keysHandler consensus.KeysHandler) *spos.ConsensusState {
	eligibleList := []string{"1", "2", "3"}

	eligibleNodesPubKeys := make(map[string]struct{})
	for _, key := range eligibleList {
		eligibleNodesPubKeys[key] = struct{}{}
	}

	rcns, _ := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		3,
		"2",
		keysHandler,
	)

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bls.SrBlock, 1)
	rthr.SetThreshold(bls.SrSignature, 3)
	rthr.SetFallbackThreshold(bls.SrBlock, 1)
	rthr.SetFallbackThreshold(bls.SrSignature, 2)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	return cns
}

func TestConsensusState__NewConsensusStateShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()
	assert.NotNil(t, cns)
}

func TestConsensusState_ResetConsensusStateShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()
	cns.RoundCanceled = true
	cns.ExtendedCalled = true
	cns.WaitingAllSignaturesTimeOut = true
	cns.ResetConsensusState()
	assert.False(t, cns.RoundCanceled)
	assert.False(t, cns.ExtendedCalled)
	assert.False(t, cns.WaitingAllSignaturesTimeOut)
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetConsensusGroup(nil)
	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("2"))
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensusState_IsSelfLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestConsensusState_IsSelfLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestConsensusState_GetLeaderShoudErrNilConsensusGroup(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetConsensusGroup(nil)

	_, err := cns.GetLeader()
	assert.Equal(t, spos.ErrNilConsensusGroup, err)
}

func TestConsensusState_GetLeaderShouldErrEmptyConsensusGroup(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetConsensusGroup(make([]string, 0))

	_, err := cns.GetLeader()
	assert.Equal(t, spos.ErrEmptyConsensusGroup, err)
}

func TestConsensusState_GetLeaderShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	leader, err := cns.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, cns.ConsensusGroup()[0], leader)
}

func TestConsensusState_GetNextConsensusGroupShouldFailWhenComputeValidatorsGroupErr(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	nodesCoord := &shardingMocks.NodesCoordinatorMock{}
	err := errors.New("error")
	nodesCoord.ComputeValidatorsGroupCalled = func(
		randomness []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) ([]nodesCoordinator.Validator, error) {
		return nil, err
	}

	_, err2 := cns.GetNextConsensusGroup([]byte(""), 0, 0, nodesCoord, 0)
	assert.Equal(t, err, err2)
}

func TestConsensusState_GetNextConsensusGroupShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	nodesCoord := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]nodesCoordinator.Validator, error) {
			defaultSelectionChances := uint32(1)
			return []nodesCoordinator.Validator{
				shardingMocks.NewValidatorMock([]byte("A"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("B"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("C"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("D"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("E"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("F"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("G"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("H"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("I"), 1, defaultSelectionChances),
			}, nil
		},
	}

	nextConsensusGroup, err := cns.GetNextConsensusGroup(nil, 0, 0, nodesCoord, 0)
	assert.Nil(t, err)
	assert.NotNil(t, nextConsensusGroup)
}

func TestConsensusState_IsConsensusDataSetShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)

	assert.True(t, cns.IsConsensusDataSet())
}

func TestConsensusState_IsConsensusDataSetShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = nil

	assert.False(t, cns.IsConsensusDataSet())
}

func TestConsensusState_IsConsensusDataEqualShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	data := []byte("consensus data")

	cns.Data = data

	assert.True(t, cns.IsConsensusDataEqual(data))
}

func TestConsensusState_IsConsensusDataEqualShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	data := []byte("consensus data")

	cns.Data = data

	assert.False(t, cns.IsConsensusDataEqual([]byte("X")))
}

func TestConsensusState_IsJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	_ = cns.SetJobDone("1", bls.SrBlock, false)
	assert.False(t, cns.IsJobDone("1", bls.SrBlock))

	_ = cns.SetJobDone("1", bls.SrSignature, true)
	assert.False(t, cns.IsJobDone("1", bls.SrBlock))

	_ = cns.SetJobDone("2", bls.SrBlock, true)
	assert.False(t, cns.IsJobDone("1", bls.SrBlock))
}

func TestConsensusState_IsJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	_ = cns.SetJobDone("1", bls.SrBlock, true)

	assert.True(t, cns.IsJobDone("1", bls.SrBlock))
}

func TestConsensusState_IsSelfJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrBlock, false)
	assert.False(t, cns.IsSelfJobDone(bls.SrBlock))

	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrSignature, true)
	assert.False(t, cns.IsSelfJobDone(bls.SrBlock))

	_ = cns.SetJobDone(cns.SelfPubKey()+"X", bls.SrBlock, true)
	assert.False(t, cns.IsSelfJobDone(bls.SrBlock))
}

func TestConsensusState_IsSelfJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrBlock, true)

	assert.True(t, cns.IsSelfJobDone(bls.SrBlock))
}

func TestConsensusState_IsCurrentSubroundFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetStatus(bls.SrBlock, spos.SsNotFinished)
	assert.False(t, cns.IsSubroundFinished(bls.SrBlock))

	cns.SetStatus(bls.SrSignature, spos.SsFinished)
	assert.False(t, cns.IsSubroundFinished(bls.SrBlock))

}

func TestConsensusState_IsCurrentSubroundFinishedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetStatus(bls.SrBlock, spos.SsFinished)
	assert.True(t, cns.IsSubroundFinished(bls.SrBlock))
}

func TestConsensusState_IsNodeSelfShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.False(t, cns.IsNodeSelf(cns.SelfPubKey()+"X"))
}

func TestConsensusState_IsNodeSelfShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.True(t, cns.IsNodeSelf(cns.SelfPubKey()))
}

func TestConsensusState_IsBlockBodyAlreadyReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Body = nil

	assert.False(t, cns.IsBlockBodyAlreadyReceived())
}

func TestConsensusState_IsBlockBodyAlreadyReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Body = &block.Body{}

	assert.True(t, cns.IsBlockBodyAlreadyReceived())
}

func TestConsensusState_IsHeaderAlreadyReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Header = nil

	assert.False(t, cns.IsHeaderAlreadyReceived())
}

func TestConsensusState_IsHeaderAlreadyReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Header = &block.Header{}

	assert.True(t, cns.IsHeaderAlreadyReceived())
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenConsensusDataNotSet(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = nil

	assert.False(t, cns.CanDoSubroundJob(bls.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenSelfJobIsDone(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)
	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrBlock, true)

	assert.False(t, cns.CanDoSubroundJob(bls.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)
	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrBlock, false)
	cns.SetStatus(bls.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanDoSubroundJob(bls.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)
	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrBlock, false)
	cns.SetStatus(bls.SrBlock, spos.SsNotFinished)

	assert.True(t, cns.CanDoSubroundJob(bls.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedFromItself(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte(cns.SelfPubKey()),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bls.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedForOtherRound(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 1, bls.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenJobIsDone(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	_ = cns.SetJobDone("1", bls.SrBlock, true)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bls.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetStatus(bls.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bls.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.True(t, cns.CanProcessReceivedMessage(cnsDta, 0, bls.SrBlock))
}

func TestConsensusState_GenerateBitmapShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	bitmapExpected := make([]byte, cns.ConsensusGroupSize()/8+1)
	selfIndexInConsensusGroup, _ := cns.SelfConsensusGroupIndex()
	bitmapExpected[selfIndexInConsensusGroup/8] |= 1 << (uint16(selfIndexInConsensusGroup) % 8)

	_ = cns.SetJobDone(cns.SelfPubKey(), bls.SrBlock, true)
	bitmap := cns.GenerateBitmap(bls.SrBlock)

	assert.Equal(t, bitmapExpected, bitmap)
}

func TestConsensusState_SetAndGetProcessingBlockShouldWork(t *testing.T) {
	t.Parallel()
	cns := internalInitConsensusState()
	cns.SetProcessingBlock(true)

	assert.Equal(t, true, cns.ProcessingBlock())
}

func TestConsensusState_IsMultiKeyLeaderInCurrentRound(t *testing.T) {
	t.Parallel()

	keysHandler := &testscommon.KeysHandlerStub{}
	cns := internalInitConsensusStateWithKeysHandler(keysHandler)
	t.Run("no managed keys from consensus group should return false", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return false
		}
		assert.False(t, cns.IsMultiKeyLeaderInCurrentRound())
	})
	t.Run("node has managed keys but no managed key is leader should return false", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("2"), pkBytes)
		}

		assert.False(t, cns.IsMultiKeyLeaderInCurrentRound())
	})
	t.Run("node has managed keys and one key is leader should return true", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("1"), pkBytes)
		}

		assert.True(t, cns.IsMultiKeyLeaderInCurrentRound())
	})
}

func TestConsensusState_IsLeaderJobDone(t *testing.T) {
	t.Parallel()

	keysHandler := &testscommon.KeysHandlerStub{}
	cns := internalInitConsensusStateWithKeysHandler(keysHandler)
	t.Run("should work", func(t *testing.T) {
		assert.False(t, cns.IsLeaderJobDone(0))
		leader, _ := cns.GetLeader()
		_ = cns.SetJobDone(leader, 0, true)
		assert.True(t, cns.IsLeaderJobDone(0))
	})
	t.Run("GetLeader errors should return false", func(t *testing.T) {
		leader, _ := cns.GetLeader()
		_ = cns.SetJobDone(leader, 0, true)
		cns.SetConsensusGroup(make([]string, 0))
		assert.False(t, cns.IsLeaderJobDone(0))
	})
}

func TestConsensusState_IsMultiKeyJobDone(t *testing.T) {
	t.Parallel()

	keysHandler := &testscommon.KeysHandlerStub{}
	cns := internalInitConsensusStateWithKeysHandler(keysHandler)
	managedKeyInConsensus := "1"
	managedKeyNotInConsensus := "managed key not in consensus group"
	t.Run("no managed keys should return true", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return false
		}

		assert.True(t, cns.IsMultiKeyJobDone(0))
	})
	t.Run("node has managed keys but no key is in consensus group should return true", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte(managedKeyNotInConsensus), pkBytes)
		}

		assert.True(t, cns.IsMultiKeyJobDone(0))
	})
	t.Run("node has managed keys and one key is in consensus group", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte(managedKeyInConsensus), pkBytes)
		}

		assert.False(t, cns.IsMultiKeyJobDone(0))
		_ = cns.SetJobDone(managedKeyInConsensus, 0, true)
		assert.True(t, cns.IsMultiKeyJobDone(0))
	})
}
