package spos

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func initConsensusDataContainer() *ConsensusCore {
	blockChain := &mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	blsPrivateKeyMock := &mock.PrivateKeyMock{}
	blsSingleSignerMock := &mock.SingleSignerMock{}
	multiSignerMock := mock.NewMultiSigner()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := &mock.RounderMock{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	return &ConsensusCore{
		blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		hasherMock,
		marshalizerMock,
		blsPrivateKeyMock,
		blsSingleSignerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
	}
}

func TestConsensusContainerValidator_ValidateNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.blockChain = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilBlockChain)
}

func TestConsensusContainerValidator_ValidateNilProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.blockProcessor = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilBlockProcessor)
}

func TestConsensusContainerValidator_ValidateNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.bootstraper = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilBlootstraper)
}

func TestConsensusContainerValidator_ValidateNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.chronologyHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilChronologyHandler)
}

func TestConsensusContainerValidator_ValidateNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.hasher = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilHasher)
}

func TestConsensusContainerValidator_ValidateNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.marshalizer = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilMarshalizer)
}

func TestConsensusContainerValidator_ValidateNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.multiSigner = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilMultiSigner)
}

func TestConsensusContainerValidator_ValidateNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.rounder = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilRounder)
}

func TestConsensusContainerValidator_ValidateNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.shardCoordinator = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilShardCoordinator)
}

func TestConsensusContainerValidator_ValidateNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.syncTimer = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilSyncTimer)
}

func TestConsensusContainerValidator_ValidateNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.validatorGroupSelector = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, err, ErrNilValidatorGroupSelector)
}
