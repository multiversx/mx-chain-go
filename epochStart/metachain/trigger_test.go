package metachain

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockEpochStartTriggerArguments() *ArgsNewMetaEpochStartTrigger {
	return &ArgsNewMetaEpochStartTrigger{
		GenesisTime: time.Time{},
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         2,
		},
		Epoch:              0,
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		Marshalizer:        &mock.MarshalizerMock{},
		Hasher:             &hashingMocks.HasherMock{},
		AppStatusHandler:   &statusHandlerMock.AppStatusHandlerStub{},
		Storage: &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) (bytes []byte, err error) {
						return []byte("hash"), nil
					},
					PutCalled: func(key, data []byte) error {
						return nil
					},
					RemoveCalled: func(key []byte) error {
						return nil
					},
					SearchFirstCalled: func(key []byte) (bytes []byte, err error) {
						return []byte("hash"), nil
					},
				}, nil
			},
		},
		DataPool: &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &testscommon.HeadersCacherStub{}
			},
		},
	}
}

func TestNewEpochStartTrigger_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	epochStartTrigger, err := NewEpochStartTrigger(nil)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilArgsNewMetaEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_NilSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings = nil

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilEpochStartSettings, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 0

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidSettingsForEpochStartTrigger))
}

func TestNewEpochStartTrigger_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.EpochStartNotifier = nil

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, epochStart.ErrNilEpochStartNotifier))
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr2(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 1
	arguments.Settings.MinRoundsBetweenEpochs = 0

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidSettingsForEpochStartTrigger))
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr3(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 4
	arguments.Settings.MinRoundsBetweenEpochs = 6

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidSettingsForEpochStartTrigger))
}

func TestNewEpochStartTrigger_MissingBootstrapUnit(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			if unitType == dataRetriever.BootstrapUnit {
				return nil, storage.ErrKeyNotFound
			}
			return &storageStubs.StorerStub{}, nil
		},
	}

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.True(t, check.IfNil(epochStartTrigger))
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestNewEpochStartTrigger_MissingMetaBlockUnit(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			if unitType == dataRetriever.MetaBlockUnit {
				return nil, storage.ErrKeyNotFound
			}
			return &storageStubs.StorerStub{}, nil
		},
	}

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.True(t, check.IfNil(epochStartTrigger))
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestNewEpochStartTrigger_ShouldOk(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.NotNil(t, epochStartTrigger)
	assert.Nil(t, err)
}

func TestTrigger_Update(t *testing.T) {
	t.Parallel()

	notifierWasCalled := false
	epoch := uint32(0)
	round := uint64(0)
	nonce := uint64(100)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	arguments.EpochStartNotifier = &mock.EpochStartNotifierStub{
		NotifyAllCalled: func(hdr data.HeaderHandler) {
			notifierWasCalled = true
		},
	}
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)

	ret := epochStartTrigger.IsEpochStart()
	assert.True(t, ret)

	epc := epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	epochStartTrigger.SetProcessed(&block.MetaBlock{
		Round:      round,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}}}}, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)
	assert.True(t, notifierWasCalled)
}

func TestTrigger_ForceEpochStartCloseToNormalEpochStartShouldNotForce(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.MinRoundsBetweenEpochs = 20
	arguments.Settings.RoundsPerEpoch = 200
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)
	epochStartTrigger.currentRound = 20

	epochStartTrigger.ForceEpochStart(201)
	assert.Equal(t, uint64(math.MaxUint64), epochStartTrigger.nextEpochStartRound)
}

func TestTrigger_ForceEpochStartUnderMinimumBetweenEpochs(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.MinRoundsBetweenEpochs = 20
	arguments.Settings.RoundsPerEpoch = 200
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)
	epochStartTrigger.currentRound = 1

	epochStartTrigger.ForceEpochStart(10)
	assert.Equal(t, uint64(arguments.Settings.MinRoundsBetweenEpochs), epochStartTrigger.nextEpochStartRound)
}

func TestTrigger_ForceEpochStartShouldOk(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.MinRoundsBetweenEpochs = 20
	arguments.Settings.RoundsPerEpoch = 200
	arguments.Epoch = epoch
	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	require.Nil(t, err)

	epochStartTrigger.currentRound = 50

	expectedRound := uint64(60)
	epochStartTrigger.ForceEpochStart(60)

	assert.Equal(t, expectedRound, epochStartTrigger.nextEpochStartRound)

	epochStartTrigger.Update(expectedRound, minimumNonceToStartEpoch)

	isEpochStart := epochStartTrigger.IsEpochStart()
	assert.True(t, isEpochStart)
}

func TestTrigger_RevertStateToBlock(t *testing.T) {
	t.Parallel()

	triggerFactory := func(arguments *ArgsNewMetaEpochStartTrigger) epochStart.TriggerHandler {
		epochStartTrigger, _ := NewEpochStartTrigger(arguments)
		return epochStartTrigger
	}
	metaHdrFactory := func(round uint64, epoch uint32, isEpochStart bool) data.MetaHeaderHandler {
		metaBlock := &block.MetaBlock{
			Round: round,
			Epoch: epoch,
		}

		if isEpochStart {
			metaBlock.EpochStart = block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}}}
		}

		return metaBlock
	}

	testTriggerRevertToEndOfEpochUpdate(t, triggerFactory, metaHdrFactory)
	testTriggerRevertBehindEpochStartBlock(t, triggerFactory, metaHdrFactory)
}

type createTriggerFunc func(arguments *ArgsNewMetaEpochStartTrigger) epochStart.TriggerHandler
type createEpochStartMetaHdrFunc func(round uint64, epoch uint32, isEpochStart bool) data.MetaHeaderHandler

func testTriggerRevertBehindEpochStartBlock(
	t *testing.T,
	createTriggerFunc createTriggerFunc,
	createEpochStartMetaHdrFunc createEpochStartMetaHdrFunc,
) {
	epoch := uint32(0)
	round := uint64(0)
	nonce := uint64(100)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch

	firstBlock := createEpochStartMetaHdrFunc(0, 0, false)
	firstBlockBuff, _ := arguments.Marshalizer.Marshal(firstBlock)

	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					return []byte("hash"), nil
				},
				PutCalled: func(key, data []byte) error {
					return nil
				},
				RemoveCalled: func(key []byte) error {
					return nil
				},
				SearchFirstCalled: func(key []byte) (bytes []byte, err error) {
					return firstBlockBuff, nil
				},
			}, nil
		},
	}

	epochStartTrigger := createTriggerFunc(arguments)

	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)

	ret := epochStartTrigger.IsEpochStart()
	assert.True(t, ret)

	epc := epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	prevMetaHdr := createEpochStartMetaHdrFunc(round-1, epoch, false)
	prevHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, prevMetaHdr)
	metaHdr := createEpochStartMetaHdrFunc(round, epoch+1, true)
	err := metaHdr.SetPrevHash(prevHash)
	require.Nil(t, err)

	epochStartTrigger.SetProcessed(metaHdr, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)

	err = epochStartTrigger.RevertStateToBlock(metaHdr)
	assert.Nil(t, err)
	assert.Equal(t, metaHdr.GetEpoch(), epochStartTrigger.Epoch())
	assert.False(t, epochStartTrigger.IsEpochStart())
	assert.Equal(t, epochStartTrigger.EpochStartRound(), metaHdr.GetRound())

	err = epochStartTrigger.RevertStateToBlock(prevMetaHdr)
	assert.Nil(t, err)
	assert.Equal(t, firstBlock.GetEpoch(), epochStartTrigger.Epoch())
	assert.False(t, epochStartTrigger.IsEpochStart())
	assert.Equal(t, epochStartTrigger.EpochStartRound(), firstBlock.GetRound())

	epochStartTrigger.Update(round, nonce)
	ret = epochStartTrigger.IsEpochStart()
	assert.True(t, ret)

	epc = epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	epochStartTrigger.SetProcessed(createEpochStartMetaHdrFunc(round, epoch+1, true), nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)
}

func testTriggerRevertToEndOfEpochUpdate(
	t *testing.T,
	createTriggerFunc createTriggerFunc,
	createEpochStartMetaHdrFunc createEpochStartMetaHdrFunc,
) {
	epoch := uint32(0)
	round := uint64(0)
	nonce := uint64(100)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	epochStartTrigger := createTriggerFunc(arguments)

	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)
	round++
	epochStartTrigger.Update(round, nonce)

	ret := epochStartTrigger.IsEpochStart()
	assert.True(t, ret)

	epc := epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	metaHdr := createEpochStartMetaHdrFunc(round, epoch+1, true)
	epochStartTrigger.SetProcessed(metaHdr, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)

	err := epochStartTrigger.RevertStateToBlock(metaHdr)
	assert.Nil(t, err)
	assert.Equal(t, metaHdr.GetEpoch(), epochStartTrigger.Epoch())
	assert.False(t, epochStartTrigger.IsEpochStart())
	assert.Equal(t, epochStartTrigger.EpochStartRound(), metaHdr.GetRound())

	epochStartTrigger.Update(round, nonce)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)

	epc = epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	epochStartTrigger.SetProcessed(createEpochStartMetaHdrFunc(round, epoch+1, true), nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)
}
