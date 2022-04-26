package storageBootstrap

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochNotifierMock "github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardStorageBootstrapper_LoadFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	wasCalledBlockchainSetHash := false
	wasCalledBlockchainSetHeader := false
	wasCalledForkDetectorAddHeader := false
	wasCalledBlockTrackerAddTrackedHeader := false
	wasCalledEpochNotifier := false
	savedLastRound := int64(0)

	marshaller := &testscommon.MarshalizerMock{}
	startRound := 4000
	hdr := &block.Header{
		Nonce:    3999,
		Round:    3999,
		RootHash: []byte("roothash"),
		ShardID:  0,
		ChainID:  []byte("1"),
	}
	hdrHash := []byte("header hash")
	hdrBytes, _ := marshaller.Marshal(hdr)
	blockStorerMock := mock.NewStorerMock()
	_ = blockStorerMock.Put(hdrHash, hdrBytes)

	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper{
			BootStorer: &mock.BoostrapStorerMock{
				GetHighestRoundCalled: func() int64 {
					return int64(startRound)
				},
				GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
					return bootstrapStorage.BootstrapData{
						LastHeader: bootstrapStorage.BootstrapHeaderInfo{
							ShardId: hdr.ShardID,
							Epoch:   hdr.Epoch,
							Nonce:   hdr.Nonce,
							Hash:    hdrHash,
						},
						HighestFinalBlockNonce: 3999,
						LastRound:              round - 1,
					}, nil
				},
				SaveLastRoundCalled: func(round int64) error {
					savedLastRound = round
					return nil
				},
			},
			ForkDetector: &mock.ForkDetectorMock{
				AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdrHash, hash)
					assert.Equal(t, process.BHProcessed, state)

					wasCalledForkDetectorAddHeader = true
					return nil
				},
			},
			BlockProcessor: &mock.BlockProcessorMock{},
			ChainHandler: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return nil
				},
				SetCurrentBlockHeaderHashCalled: func(bytes []byte) {
					assert.Equal(t, hdrHash, bytes)
					wasCalledBlockchainSetHash = true
				},
				SetCurrentBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, rootHash []byte) error {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdr.RootHash, rootHash)
					wasCalledBlockchainSetHeader = true

					return nil
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					if wasCalledBlockchainSetHeader {
						return hdr
					}

					return nil
				},
			},
			Marshalizer: &testscommon.MarshalizerMock{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return blockStorerMock
				},
			},
			Uint64Converter:     testscommon.NewNonceHashConverterMock(),
			BootstrapRoundIndex: uint64(startRound - 1),
			ShardCoordinator:    testscommon.NewMultiShardsCoordinatorMock(1),
			NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
			EpochStartTrigger:   &mock.EpochStartTriggerStub{},
			BlockTracker: &mock.BlockTrackerMock{
				AddTrackedHeaderCalled: func(header data.HeaderHandler, hash []byte) {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdrHash, hash)

					wasCalledBlockTrackerAddTrackedHeader = true
				},
			},
			ChainID:                      string(hdr.ChainID),
			ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
			MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
			EpochNotifier: &epochNotifierMock.EpochNotifierStub{
				CheckEpochCalled: func(header data.HeaderHandler) {
					assert.Equal(t, hdr, header)
					wasCalledEpochNotifier = true
				},
			},
		},
	}

	ssb, err := NewShardStorageBootstrapper(args)
	require.Nil(t, err)

	err = ssb.LoadFromStorage()
	assert.Nil(t, err)
	assert.True(t, wasCalledBlockchainSetHash)
	assert.True(t, wasCalledBlockchainSetHeader)
	assert.True(t, wasCalledForkDetectorAddHeader)
	assert.True(t, wasCalledBlockTrackerAddTrackedHeader)
	assert.Equal(t, int64(3999), savedLastRound)
	assert.True(t, wasCalledEpochNotifier)
}
