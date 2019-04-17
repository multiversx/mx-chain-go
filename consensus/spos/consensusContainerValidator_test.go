package spos

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initConsensusDataContainer() *ConsensusDataContainer {
	blockChain := &mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	multiSignerMock := mock.NewMultiSigner()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := &mock.RounderMock{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	return &ConsensusDataContainer{
		blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
	}
}

func TestConsensusContainerValidator_ValidateNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.blockChain = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilBlockChain)
}

func TestConsensusContainerValidator_ValidateNilProcessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.blockProcessor = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilBlockProcessor)
}

func TestConsensusContainerValidator_ValidateNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.bootstraper = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilBlootstraper)
}

func TestConsensusContainerValidator_ValidateNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.chronologyHandler = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilChronologyHandler)
}

func TestConsensusContainerValidator_ValidateNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.hasher = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilHasher)
}

func TestConsensusContainerValidator_ValidateNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.marshalizer = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilMarshalizer)
}

func TestConsensusContainerValidator_ValidateNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.multiSigner = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilMultiSigner)
}

func TestConsensusContainerValidator_ValidateNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.rounder = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilRounder)
}

func TestConsensusContainerValidator_ValidateNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.shardCoordinator = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilShardCoordinator)
}

func TestConsensusContainerValidator_ValidateNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.syncTimer = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilSyncTimer)
}

func TestConsensusContainerValidator_ValidateNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusContainerValidator := &ConsensusContainerValidator{}
	container := initConsensusDataContainer()
	container.validatorGroupSelector = nil

	err := consensusContainerValidator.ValidateConsensusDataContainer(container)

	assert.Equal(t, err, ErrNilValidatorGroupSelector)
}
