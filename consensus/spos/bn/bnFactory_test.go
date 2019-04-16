package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func initFactory() bn.Factory {
	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}

	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	validatorGroupSelector := mock.ValidatorGroupSelectorMock{}
	worker := initWorker()

	fct, _ := bn.NewFactory(
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

	return fct
}

func TestFactory_GetMessageTypeName(t *testing.T) {
	t.Parallel()

	r := bn.GetStringValue(bn.MtBlockBody)
	assert.Equal(t, "(BLOCK_BODY)", r)

	r = bn.GetStringValue(bn.MtBlockHeader)
	assert.Equal(t, "(BLOCK_HEADER)", r)

	r = bn.GetStringValue(bn.MtCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)

	r = bn.GetStringValue(bn.MtBitmap)
	assert.Equal(t, "(BITMAP)", r)

	r = bn.GetStringValue(bn.MtCommitment)
	assert.Equal(t, "(COMMITMENT)", r)

	r = bn.GetStringValue(bn.MtSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bn.GetStringValue(bn.MtUnknown)
	assert.Equal(t, "(UNKNOWN)", r)

	r = bn.GetStringValue(spos.MessageType(-1))
	assert.Equal(t, "Undefined message type", r)
}

func TestFactory_NewFactoryNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	blockChain := mock.BlockChainMock{}
	blockProcessorMock := mock.InitBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	consensusState := mock.InitConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := mock.InitMultiSignerMock()
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
	t.Parallel()

	fct := *initFactory()

	assert.NotNil(t, fct)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()

	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateStartRoundSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundStartRoundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateStartRoundSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundBlockShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateBlockSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundBlockShouldFailWhenNewSubroundBlockFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateBlockSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundCommitmentHashShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateCommitmentHashSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundCommitmentHashShouldFailWhenNewSubroundCommitmentHashFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateCommitmentHashSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundBitmapShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateBitmapSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundBitmapShouldFailWhenNewSubroundBitmapFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateBitmapSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundCommitmentShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateCommitmentSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundCommitmentShouldFailWhenNewSubroundCommitmentFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateCommitmentSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundSignatureShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateSignatureSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundSignatureShouldFailWhenNewSubroundSignatureFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateSignatureSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundEndRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().SetConsensusStateChangedChannels(nil)

	err := fct.GenerateEndRoundSubround()
	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundEndRoundShouldFailWhenNewSubroundEndRoundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.SetSyncTimer(nil)

	err := fct.GenerateEndRoundSubround()
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	t.Parallel()

	fct := *initFactory()

	subroundHandlers := 0

	chrm := &mock.ChronologyHandlerMock{}
	chrm.AddSubroundCalled = func(subroundHandler consensus.SubroundHandler) {
		subroundHandlers++
	}

	fct.SetChronologyHandler(chrm)

	fct.GenerateSubrounds()

	assert.Equal(t, 7, subroundHandlers)
}
