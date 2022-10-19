package sposFactory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestGetSubroundsFactoryV2_BlsNilConsensusCoreShouldErr(t *testing.T) {
	t.Parallel()

	worker := &mock.SposWorkerMock{}
	consensusType := consensus.BlsConsensusType
	statusHandler := statusHandlerMock.NewAppStatusHandlerMock()
	chainID := []byte("chain-id")
	indexer := &testscommon.OutportStub{}
	sfV2, err := sposFactory.GetSubroundsFactoryV2(
		nil,
		&spos.ConsensusState{},
		worker,
		consensusType,
		statusHandler,
		indexer,
		chainID,
		currentPid,
	)

	assert.Nil(t, sfV2)
	assert.Equal(t, spos.ErrNilConsensusCore, err)
}

func TestGetSubroundsFactoryV2_BlsNilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	consensusCore := mock.InitConsensusCore()
	worker := &mock.SposWorkerMock{}
	consensusType := consensus.BlsConsensusType
	chainID := []byte("chain-id")
	indexer := &testscommon.OutportStub{}
	sfV2, err := sposFactory.GetSubroundsFactoryV2(
		consensusCore,
		&spos.ConsensusState{},
		worker,
		consensusType,
		nil,
		indexer,
		chainID,
		currentPid,
	)

	assert.Nil(t, sfV2)
	assert.Equal(t, spos.ErrNilAppStatusHandler, err)
}

func TestGetSubroundsFactoryV2_BlsShouldWork(t *testing.T) {
	t.Parallel()

	consensusCore := mock.InitConsensusCore()
	worker := &mock.SposWorkerMock{}
	consensusType := consensus.BlsConsensusType
	statusHandler := statusHandlerMock.NewAppStatusHandlerMock()
	chainID := []byte("chain-id")
	indexer := &testscommon.OutportStub{}
	sfV2, err := sposFactory.GetSubroundsFactoryV2(
		consensusCore,
		&spos.ConsensusState{},
		worker,
		consensusType,
		statusHandler,
		indexer,
		chainID,
		currentPid,
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sfV2))
}

func TestGetSubroundsFactoryV2_InvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	consensusType := "invalid"
	sfV2, err := sposFactory.GetSubroundsFactoryV2(
		nil,
		nil,
		nil,
		consensusType,
		nil,
		nil,
		nil,
		currentPid,
	)

	assert.Nil(t, sfV2)
	assert.Equal(t, sposFactory.ErrInvalidConsensusType, err)
}
