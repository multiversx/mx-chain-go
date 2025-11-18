package metachain

import (
	"errors"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
)

func createMockEpochStartTriggerArguments() *ArgsNewMetaEpochStartTrigger {
	return &ArgsNewMetaEpochStartTrigger{
		GenesisTime:        time.Time{},
		Settings:           &config.EpochStartConfig{},
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
		},
		ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
			CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
				return config.ChainParametersByEpochConfig{
					RoundsPerEpoch:         2,
					MinRoundsBetweenEpochs: 1,
				}
			},
		},
	}
}

func requireFlags(t *testing.T, expectedVal bool, flags ...*bool) {
	for _, flag := range flags {
		require.Equal(t, expectedVal, *flag)
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

func TestNewEpochStartTrigger_NilChainParametersHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.ChainParametersHandler = nil

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, process.ErrNilChainParametersHandler))
}

func TestNewEpochStartTrigger_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.EpochStartNotifier = nil

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, epochStart.ErrNilEpochStartNotifier))
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

func TestNewEpochStartTrigger_UpdateRoundAndSetEpochChange(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	require.NotNil(t, epochStartTrigger)
	require.Nil(t, err)

	epoch := uint32(0)
	round := uint64(2)
	nonce := uint64(100)

	shouldProposeEpochChange := epochStartTrigger.ShouldProposeEpochChange(round, nonce)
	require.True(t, shouldProposeEpochChange)

	epochStartTrigger.SetEpochChange(round)
	currentEpoch := epochStartTrigger.Epoch()
	require.Equal(t, epoch+1, currentEpoch)
	require.True(t, epochStartTrigger.IsEpochStart())
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
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	epochStartTrigger.ForceEpochStart(201)
	assert.Equal(t, uint64(math.MaxUint64), epochStartTrigger.nextEpochStartRound)
}

func TestTrigger_ForceEpochStartUnderMinimumBetweenEpochs(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.ChainParametersHandler = &chainParameters.ChainParametersHandlerStub{
		ChainParametersForEpochCalled: func(uint32) (config.ChainParametersByEpochConfig, error) {
			return config.ChainParametersByEpochConfig{
				MinRoundsBetweenEpochs: 20,
				RoundsPerEpoch:         200,
			}, nil
		},
	}
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	epochStartTrigger.ForceEpochStart(10)
	assert.Equal(t, uint64(20), epochStartTrigger.nextEpochStartRound)
}

func TestTrigger_ForceEpochStartShouldOk(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	arguments := createMockEpochStartTriggerArguments()
	arguments.ChainParametersHandler = &chainParameters.ChainParametersHandlerStub{
		ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
			return config.ChainParametersByEpochConfig{
				MinRoundsBetweenEpochs: 20,
				RoundsPerEpoch:         200,
			}, nil
		},
	}

	arguments.Epoch = epoch
	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	require.Nil(t, err)

	expectedRound := uint64(60)
	epochStartTrigger.ForceEpochStart(60)

	assert.Equal(t, expectedRound, epochStartTrigger.nextEpochStartRound)

	epochStartTrigger.Update(expectedRound, minimumNonceToStartEpoch)

	isEpochStart := epochStartTrigger.IsEpochStart()
	assert.True(t, isEpochStart)
}

func TestTrigger_LastCommitedMetaEpochStartBlock(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartTriggerArguments()
	et, _ := NewEpochStartTrigger(args)

	epoch := uint32(37)

	epochStartNonce := uint64(100)
	epochStartRound := uint64(101)
	ecpohStartTimeStamp := uint64(102)

	epochStartMetaHdr := &block.MetaBlock{
		Epoch:     epoch,
		Nonce:     epochStartNonce,
		Round:     epochStartRound,
		TimeStamp: ecpohStartTimeStamp,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}},
		},
	}

	nonce := uint64(200)
	round := uint64(201)
	timeStamp := uint64(202)

	metaHdr := &block.MetaBlock{
		Epoch:     epoch,
		Nonce:     nonce,
		Round:     round,
		TimeStamp: timeStamp,
	}

	et.SetProcessed(epochStartMetaHdr, nil)
	et.SetProcessed(metaHdr, nil)

	lastCommitedEpochStartBlock, err := et.LastCommitedEpochStartHdr()
	require.Nil(t, err)
	require.Equal(t, epochStartMetaHdr, lastCommitedEpochStartBlock)
}

func TestTrigger_UpdateRevertToEndOfEpochUpdate(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	round := uint64(0)
	nonce := uint64(100)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
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

	metaHdr := &block.MetaBlock{
		Round: round,
		Epoch: epoch + 1,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}},
			Economics: block.Economics{
				TotalSupply:         big.NewInt(0),
				TotalToDistribute:   big.NewInt(0),
				TotalNewlyMinted:    big.NewInt(0),
				RewardsPerBlock:     big.NewInt(0),
				NodePrice:           big.NewInt(0),
				PrevEpochStartRound: 0,
			}}}
	epochStartTrigger.SetProcessed(metaHdr, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)

	err := epochStartTrigger.RevertStateToBlock(metaHdr)
	assert.Nil(t, err)
	assert.Equal(t, metaHdr.Epoch, epochStartTrigger.Epoch())
	assert.False(t, epochStartTrigger.IsEpochStart())
	assert.Equal(t, epochStartTrigger.currEpochStartRound, metaHdr.Round)

	epochStartTrigger.Update(round, nonce)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)

	epc = epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	epochStartTrigger.SetProcessed(&block.MetaBlock{
		Round:      round,
		Epoch:      epoch + 1,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}}}}, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)
}

func TestTrigger_RevertBehindEpochStartBlock(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	round := uint64(0)
	nonce := uint64(100)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	firstBlock := &block.MetaBlock{}
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

	prevMetaHdr := &block.MetaBlock{
		Round: round - 1,
		Epoch: epoch,
	}

	prevHash, _ := core.CalculateHash(epochStartTrigger.marshaller, epochStartTrigger.hasher, prevMetaHdr)
	metaHdr := &block.MetaBlock{
		Round:    round,
		Epoch:    epoch + 1,
		PrevHash: prevHash,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}},
			Economics: block.Economics{
				TotalSupply:         big.NewInt(0),
				TotalToDistribute:   big.NewInt(0),
				TotalNewlyMinted:    big.NewInt(0),
				RewardsPerBlock:     big.NewInt(0),
				NodePrice:           big.NewInt(0),
				PrevEpochStartRound: 0,
			}}}
	epochStartTrigger.SetProcessed(metaHdr, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)

	err := epochStartTrigger.RevertStateToBlock(metaHdr)
	assert.Nil(t, err)
	assert.Equal(t, metaHdr.Epoch, epochStartTrigger.Epoch())
	assert.False(t, epochStartTrigger.IsEpochStart())
	assert.Equal(t, epochStartTrigger.currEpochStartRound, metaHdr.Round)

	err = epochStartTrigger.RevertStateToBlock(prevMetaHdr)
	assert.Nil(t, err)
	assert.Equal(t, firstBlock.Epoch, epochStartTrigger.Epoch())
	assert.False(t, epochStartTrigger.IsEpochStart())
	assert.Equal(t, epochStartTrigger.currEpochStartRound, firstBlock.Round)

	epochStartTrigger.Update(round, nonce)
	ret = epochStartTrigger.IsEpochStart()
	assert.True(t, ret)

	epc = epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	epochStartTrigger.SetProcessed(&block.MetaBlock{
		Round:      round,
		Epoch:      epoch + 1,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}}}}, nil)
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)
}

func TestTrigger_SetProcessedHeaderV3(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartTriggerArguments()

	wasNotifyAllCalled := false
	wasNotifyAllPrepareCalled := false
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		NotifyAllCalled: func(hdr data.HeaderHandler) {
			wasNotifyAllCalled = true
		},
		NotifyAllPrepareCalled: func(hdr data.HeaderHandler, body data.BodyHandler) {
			wasNotifyAllPrepareCalled = true
		},
	}

	epochStartTrigger, _ := NewEpochStartTrigger(args)

	header := &block.MetaBlockV3{Nonce: 4, EpochStart: block.EpochStart{LastFinalizedHeaders: make([]block.EpochStartShardData, 1)}}
	epochStartTrigger.SetProcessed(header, &block.Body{})

	require.True(t, wasNotifyAllCalled)
	require.False(t, wasNotifyAllPrepareCalled)
	require.False(t, epochStartTrigger.isEpochStart)
}

func TestTrigger_SetProposed(t *testing.T) {

	args := createMockEpochStartTriggerArguments()

	wasNotifyAllCalled := false
	wasNotifyAllPrepareCalled := false
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		NotifyAllCalled: func(hdr data.HeaderHandler) {
			wasNotifyAllCalled = true
		},
		NotifyAllPrepareCalled: func(hdr data.HeaderHandler, body data.BodyHandler) {
			wasNotifyAllPrepareCalled = true
		},
	}

	var wasStateSaved bool
	storer := &storageStubs.StorerStub{
		PutCalled: func(key, data []byte) error {
			wasStateSaved = true
			return nil
		},
	}
	args.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storer, nil
		},
	}

	epochStartTrigger, _ := NewEpochStartTrigger(args)
	wasStateSaved = false

	allFlags := []*bool{&wasNotifyAllCalled, &wasStateSaved, &wasNotifyAllPrepareCalled, &epochStartTrigger.isEpochStart}
	t.Run("not meta block", func(t *testing.T) {
		epochStartTrigger.SetProposed(&block.HeaderV3{Nonce: 4}, &block.Body{})
		requireFlags(t, false, allFlags...)
	})
	t.Run("not meta header v3", func(t *testing.T) {
		epochStartTrigger.SetProposed(&block.MetaBlock{Nonce: 4}, &block.Body{})
		requireFlags(t, false, allFlags...)
	})
	t.Run("not epoch change proposed", func(t *testing.T) {
		epochStartTrigger.SetProposed(&block.MetaBlockV3{Nonce: 4, EpochChangeProposed: false}, &block.Body{})
		requireFlags(t, false, allFlags...)
	})
	t.Run("should work", func(t *testing.T) {
		epochStartTrigger.SetProposed(&block.MetaBlockV3{Nonce: 4, EpochChangeProposed: true}, &block.Body{})
		allFlags = allFlags[1:] // remove wasNotifyAllCalled flag
		requireFlags(t, true, allFlags...)
		require.False(t, wasNotifyAllCalled)
	})
}
