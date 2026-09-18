package v2_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

const competingParentNonce = uint64(7)

var (
	competingParentHash = []byte("current-meta-parent")
	competingChildHash  = []byte("candidate-meta-child")
)

func configureMetaCompetingParent(
	container *spos.ConsensusCore,
	evidenceAtQuery int,
	childProofPresent bool,
) *int {
	currentParentHash := []byte("current-meta-grandparent")
	currentParent := &block.MetaBlockV3{
		Epoch:    1,
		Nonce:    competingParentNonce,
		Round:    10,
		PrevHash: currentParentHash,
	}
	siblingHash := []byte("preferred-meta-parent")
	siblingProof := &block.HeaderProof{
		HeaderHash:    siblingHash,
		HeaderEpoch:   1,
		HeaderNonce:   competingParentNonce,
		HeaderShardId: core.MetachainShardId,
		HeaderRound:   9,
	}

	container.SetShardCoordinator(mock.ShardCoordinatorMock{ShardID: core.MetachainShardId})
	container.SetBlockchain(&testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderAndHashCalled: func() (data.HeaderHandler, []byte) {
			return currentParent, competingParentHash
		},
	})
	container.SetHeadersPool(&dataRetrieverMock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if !bytes.Equal(hash, siblingHash) {
				return nil, errors.New("header not found")
			}

			return &block.MetaBlockV3{
				Nonce:    competingParentNonce,
				Round:    siblingProof.HeaderRound,
				PrevHash: currentParentHash,
			}, nil
		},
	})
	container.SetEnableEpochsHandler(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SupernovaFlag
		},
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return flag == common.SupernovaFlag
		},
	})
	container.SetEnableRoundsHandler(&testscommon.EnableRoundsHandlerStub{
		IsFlagEnabledCalled: func(flag common.EnableRoundFlag) bool {
			return flag == common.SupernovaRoundFlag
		},
		IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, _ uint64) bool {
			return flag == common.SupernovaRoundFlag
		},
	})

	parentQueries := 0
	container.SetEquivalentProofsPool(&dataRetrieverTests.ProofsPoolMock{
		HasProofCalled: func(_ uint32, hash []byte) bool {
			return childProofPresent && bytes.Equal(hash, competingChildHash)
		},
		GetProofByNonceCalled: func(nonce uint64, _ uint32) (data.HeaderProofHandler, error) {
			if nonce != competingParentNonce {
				return nil, errors.New("proof not found")
			}

			parentQueries++
			if parentQueries < evidenceAtQuery {
				return nil, errors.New("proof not found")
			}

			return siblingProof, nil
		},
	})

	return &parentQueries
}

func competingMetaChild() *block.MetaBlockV3 {
	return &block.MetaBlockV3{
		Epoch:    1,
		Nonce:    competingParentNonce + 1,
		Round:    11,
		PrevHash: competingParentHash,
	}
}

func TestSubroundSignature_CompetingParentGuard(t *testing.T) {
	t.Parallel()

	t.Run("late parent evidence prevents signing", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		parentQueries := configureMetaCompetingParent(container, 1, false)
		broadcastCalled := false
		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(_ *consensus.Message) error {
				broadcastCalled = true
				return nil
			},
		})

		sr := initSubroundSignatureWithContainer(container)
		sr.SetHeader(competingMetaChild())
		sr.SetData(competingChildHash)
		leader, err := sr.GetLeader()
		require.NoError(t, err)
		sr.SetSelfPubKey(leader)

		require.False(t, sr.DoSignatureJob())
		require.False(t, broadcastCalled)
		require.Equal(t, 1, *parentQueries)
	})

	t.Run("existing child proof is honored before parent evidence", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		parentQueries := configureMetaCompetingParent(container, 1, true)
		sr := initSubroundSignatureWithContainer(container)
		sr.SetHeader(competingMetaChild())
		sr.SetData(competingChildHash)

		require.True(t, sr.DoSignatureJob())
		require.Zero(t, *parentQueries)
	})
}

func TestSubroundEndRound_CompetingParentGuardIsRechecked(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		evidenceAtQuery int
		expectedError   error
	}{
		{name: "before aggregation", evidenceAtQuery: 1},
		{name: "during aggregation", evidenceAtQuery: 2},
		{name: "before proof insertion", evidenceAtQuery: 3, expectedError: v2.ErrProofAlreadyPropagated},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			container := consensusMocks.InitConsensusCore()
			parentQueries := configureMetaCompetingParent(container, testCase.evidenceAtQuery, false)
			broadcastCalled := false
			container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
				BroadcastEquivalentProofCalled: func(_ data.HeaderProofHandler, _ []byte) error {
					broadcastCalled = true
					return nil
				},
			})

			sr := initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
			sr.SetHeader(competingMetaChild())
			sr.SetData(competingChildHash)
			for _, publicKey := range sr.ConsensusGroup() {
				require.NoError(t, sr.SetJobDone(publicKey, bls.SrSignature, true))
			}

			sent, err := sr.SendProof()
			require.False(t, sent)
			if testCase.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, testCase.expectedError)
			}
			require.False(t, broadcastCalled)
			require.Equal(t, testCase.evidenceAtQuery, *parentQueries)
		})
	}
}
