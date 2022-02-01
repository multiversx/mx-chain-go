package metachain

import (
	"errors"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
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
		Storage: &mock.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
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
				}
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

	arguments.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
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
			}
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
