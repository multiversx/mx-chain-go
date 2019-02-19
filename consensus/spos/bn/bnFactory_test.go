package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func initFactory() (*bn.Factory, error) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	return fct, err
}

func TestFactory_NewFactoryNilBlockchainShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		nil,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestFactory_NewFactoryNilBlockProcessorShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		nil,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestFactory_NewFactoryNilBootstraperShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		nil,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilBlootstraper)
}

func TestFactory_NewFactoryNilChronologyHandlerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		nil,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilChronologyHandler)
}

func TestFactory_NewFactoryNilConsensusStateShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		nil,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestFactory_NewFactoryNilHasherShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		nil,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestFactory_NewFactoryNilMarshalizerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		nil,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestFactory_NewFactoryNilMultiSignerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		nil,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestFactory_NewFactoryNilRounderShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		nil,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestFactory_NewFactoryNilShardCoordinatorShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		nil,
		syncTimerMock,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilShardCoordinator)
}

func TestFactory_NewFactoryNilSyncTimerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		nil,
		validatorGroupSelector,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestFactory_NewFactoryNilValidatorGroupSelectorShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	worker := initWorker()

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		nil,
		worker,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilValidatorGroupSelector)
}

func TestFactory_NewFactoryNilWorkerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	chronologyHandlerMock := initChronologyHandlerMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	fct, err := bn.NewFactory(
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		chronologyHandlerMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, err, spos.ErrNilWorker)
}

func TestFactory_NewFactoryShouldWork(t *testing.T) {
	fct, err := initFactory()

	assert.Nil(t, err)
	assert.NotNil(t, fct)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateStartRoundSubround())
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundStartRoundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateStartRoundSubround())
}

func TestFactory_GenerateSubroundBlockShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateBlockSubround())
}

func TestFactory_GenerateSubroundBlockShouldFailWhenNewSubroundBlockFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateBlockSubround())
}

func TestFactory_GenerateSubroundCommitmentHashShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateCommitmentHashSubround())
}

func TestFactory_GenerateSubroundCommitmentHashShouldFailWhenNewSubroundCommitmentHashFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateCommitmentHashSubround())
}

func TestFactory_GenerateSubroundBitmapShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateBitmapSubround())
}

func TestFactory_GenerateSubroundBitmapShouldFailWhenNewSubroundBitmapFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateBitmapSubround())
}

func TestFactory_GenerateSubroundCommitmentShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateCommitmentSubround())
}

func TestFactory_GenerateSubroundCommitmentShouldFailWhenNewSubroundCommitmentFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateCommitmentSubround())
}

func TestFactory_GenerateSubroundSignatureShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateSignatureSubround())
}

func TestFactory_GenerateSubroundSignatureShouldFailWhenNewSubroundSignatureFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateSignatureSubround())
}

func TestFactory_GenerateSubroundEndRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.Worker().ConsensusStateChangedChannel = nil

	assert.False(t, fct.GenerateEndRoundSubround())
}

func TestFactory_GenerateSubroundEndRoundShouldFailWhenNewSubroundEndRoundFail(t *testing.T) {
	fct, _ := initFactory()
	fct.SetSyncTimer(nil)

	assert.False(t, fct.GenerateEndRoundSubround())
}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	fct, _ := initFactory()

	subroundHandlers := 0

	chrm := &mock.ChronologyHandlerMock{}
	chrm.AddSubroundCalled = func(subroundHandler chronology.SubroundHandler) {
		subroundHandlers++
	}

	fct.SetChronologyHandler(chrm)

	fct.GenerateSubrounds()

	assert.Equal(t, 7, subroundHandlers)
}
