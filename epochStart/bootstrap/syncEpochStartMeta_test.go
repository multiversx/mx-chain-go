package bootstrap

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEpochStartMetaSyncer_NilsShouldError(t *testing.T) {
	t.Parallel()

	args := getEpochStartSyncerArgs()
	args.CoreComponentsHolder = nil
	ess, err := NewEpochStartMetaSyncer(args)
	assert.True(t, check.IfNil(ess))
	assert.Equal(t, epochStart.ErrNilCoreComponentsHolder, err)

	args = getEpochStartSyncerArgs()
	args.CryptoComponentsHolder = nil
	ess, err = NewEpochStartMetaSyncer(args)
	assert.True(t, check.IfNil(ess))
	assert.Equal(t, epochStart.ErrNilCryptoComponentsHolder, err)

	args = getEpochStartSyncerArgs()
	args.HeaderIntegrityVerifier = nil
	ess, err = NewEpochStartMetaSyncer(args)
	assert.True(t, check.IfNil(ess))
	assert.Equal(t, epochStart.ErrNilHeaderIntegrityVerifier, err)

	args = getEpochStartSyncerArgs()
	args.MetaBlockProcessor = nil
	ess, err = NewEpochStartMetaSyncer(args)
	assert.True(t, check.IfNil(ess))
	assert.Equal(t, epochStart.ErrNilMetablockProcessor, err)
}

func TestNewEpochStartMetaSyncer_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getEpochStartSyncerArgs()
	ess, err := NewEpochStartMetaSyncer(args)
	require.NoError(t, err)
	require.False(t, check.IfNil(ess))
}

func TestEpochStartMetaSyncer_SyncEpochStartMetaRegisterMessengerProcessorFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")

	args := getEpochStartSyncerArgs()
	messenger := &p2pmocks.MessengerStub{
		RegisterMessageProcessorCalled: func(_ string, _ string, _ p2p.MessageProcessor) error {
			return expectedErr
		},
	}
	args.Messenger = messenger
	ess, _ := NewEpochStartMetaSyncer(args)

	mb, err := ess.SyncEpochStartMeta(time.Second)
	require.Equal(t, expectedErr, err)
	require.Nil(t, mb)
}

func TestEpochStartMetaSyncer_SyncEpochStartMetaProcessorFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")

	args := getEpochStartSyncerArgs()
	messenger := &p2pmocks.MessengerStub{
		ConnectedPeersCalled: func() []core.PeerID {
			return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
		},
	}
	args.Messenger = messenger
	ess, _ := NewEpochStartMetaSyncer(args)

	mbIntercProc := &mock.MetaBlockInterceptorProcessorStub{
		GetEpochStartMetaBlockCalled: func() (data.MetaHeaderHandler, error) {
			return nil, expectedErr
		},
	}
	ess.SetEpochStartMetaBlockInterceptorProcessor(mbIntercProc)

	mb, err := ess.SyncEpochStartMeta(time.Second)
	require.Equal(t, expectedErr, err)
	require.Nil(t, mb)
}

func TestEpochStartMetaSyncer_SyncEpochStartMetaShouldWork(t *testing.T) {
	t.Parallel()

	expectedMb := &block.MetaBlock{Nonce: 37}

	args := getEpochStartSyncerArgs()
	messenger := &p2pmocks.MessengerStub{
		ConnectedPeersCalled: func() []core.PeerID {
			return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
		},
	}
	args.Messenger = messenger
	ess, _ := NewEpochStartMetaSyncer(args)

	mbIntercProc := &mock.MetaBlockInterceptorProcessorStub{
		GetEpochStartMetaBlockCalled: func() (data.MetaHeaderHandler, error) {
			return expectedMb, nil
		},
	}
	ess.SetEpochStartMetaBlockInterceptorProcessor(mbIntercProc)

	mb, err := ess.SyncEpochStartMeta(time.Second)
	require.NoError(t, err)
	require.Equal(t, expectedMb, mb)
}

func getEpochStartSyncerArgs() ArgsNewEpochStartMetaSyncer {
	return ArgsNewEpochStartMetaSyncer{
		CoreComponentsHolder: &mock.CoreComponentsMock{
			IntMarsh:            &mock.MarshalizerMock{},
			Marsh:               &mock.MarshalizerMock{},
			Hash:                &hashingMocks.HasherMock{},
			UInt64ByteSliceConv: &mock.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:      testscommon.NewPubkeyConverterMock(32),
			PathHdl:             &testscommon.PathManagerStub{},
			ChainIdCalled: func() string {
				return "chain-ID"
			},
		},
		CryptoComponentsHolder: &mock.CryptoComponentsMock{
			PubKey:   &cryptoMocks.PublicKeyStub{},
			BlockSig: &cryptoMocks.SignerStub{},
			TxSig:    &cryptoMocks.SignerStub{},
			BlKeyGen: &cryptoMocks.KeyGenStub{},
			TxKeyGen: &cryptoMocks.KeyGenStub{},
		},
		RequestHandler:   &testscommon.RequestHandlerStub{},
		Messenger:        &p2pmocks.MessengerStub{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		EconomicsData:    &economicsmocks.EconomicsHandlerStub{},
		WhitelistHandler: &testscommon.WhiteListHandlerStub{},
		StartInEpochConfig: config.EpochStartConfig{
			MinNumConnectedPeersToStart:       2,
			MinNumOfPeersToConsiderBlockValid: 2,
		},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		MetaBlockProcessor:      &mock.EpochStartMetaBlockProcessorStub{},
	}
}
