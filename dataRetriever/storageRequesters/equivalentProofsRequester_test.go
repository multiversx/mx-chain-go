package storagerequesters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	chainStorage "github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func createMockArgEquivalentProofsRequester() ArgEquivalentProofsRequester {
	return ArgEquivalentProofsRequester{
		Messenger:                &mock.MessageHandlerStub{},
		ResponseTopicName:        "",
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
		NonceConverter: &mock.Uint64ByteSliceConverterMock{
			ToByteSliceCalled: func(u uint64) []byte {
				return make([]byte, 0)
			},
			ToUint64Called: func(bytes []byte) (uint64, error) {
				return 0, nil
			},
		},
		Storage:    &genericMocks.ChainStorerMock{},
		Marshaller: &mock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		},
	}
}

func TestNewEquivalentProofsRequester(t *testing.T) {
	t.Parallel()

	t.Run("nil Messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Messenger = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilMessenger, err)
		require.Nil(t, req)
	})
	t.Run("nil ManualEpochStartNotifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.ManualEpochStartNotifier = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)
		require.Nil(t, req)
	})
	t.Run("nil ChanGracefullyClose should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.ChanGracefullyClose = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilGracefullyCloseChannel, err)
		require.Nil(t, req)
	})
	t.Run("nil NonceConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.NonceConverter = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
		require.Nil(t, req)
	})
	t.Run("nil Storage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Storage = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilStore, err)
		require.Nil(t, req)
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Marshaller = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		require.Nil(t, req)
	})
	t.Run("nil EnableEpochsHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.EnableEpochsHandler = nil
		req, err := NewEquivalentProofsRequester(args)
		require.Equal(t, dataRetriever.ErrNilEnableEpochsHandler, err)
		require.Nil(t, req)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		req, err := NewEquivalentProofsRequester(createMockArgEquivalentProofsRequester())
		require.NoError(t, err)
		require.NotNil(t, req)
	})
}

func TestEquivalentProofsRequester_IsInterfaceNil(t *testing.T) {
	var req *equivalentProofsRequester
	require.True(t, req.IsInterfaceNil())

	req, _ = NewEquivalentProofsRequester(createMockArgEquivalentProofsRequester())
	require.False(t, req.IsInterfaceNil())
}

func TestEquivalentProofsRequester_RequestDataFromHash(t *testing.T) {
	t.Parallel()

	t.Run("invalid key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromHash([]byte("invalid key"), 0)
		require.Error(t, err)
	})
	t.Run("GetStorer error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Storage = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return nil, expectedErr
			},
		}
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromHash([]byte(common.GetEquivalentProofHashShardKey([]byte("hash"), 1)), 0)
		require.Equal(t, expectedErr, err)
	})
	t.Run("SearchFirst error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Storage = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return &storage.StorerStub{
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromHash([]byte(common.GetEquivalentProofHashShardKey([]byte("hash"), 1)), 0)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work and send to self", func(t *testing.T) {
		t.Parallel()

		providedBuff := []byte("provided buff")
		args := createMockArgEquivalentProofsRequester()
		args.Storage = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return &storage.StorerStub{
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return providedBuff, nil
					},
				}, nil
			},
		}
		wasSendToConnectedPeerCalled := false
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				wasSendToConnectedPeerCalled = true
				require.Equal(t, string(providedBuff), string(buff))
				return nil
			},
		}
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromHash([]byte(common.GetEquivalentProofHashShardKey([]byte("hash"), 1)), 0)
		require.NoError(t, err)
		require.True(t, wasSendToConnectedPeerCalled)
	})
}

func TestEquivalentProofsRequester_RequestDataFromNonce(t *testing.T) {
	t.Parallel()

	t.Run("invalid key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromNonce([]byte("invalid key"), 0)
		require.Error(t, err)
	})
	t.Run("getStorerForShard error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Storage = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return nil, expectedErr
			},
		}
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromNonce([]byte(common.GetEquivalentProofNonceShardKey(123, 1)), 0)
		require.Equal(t, expectedErr, err)
	})
	t.Run("SearchFirst error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsRequester()
		args.Storage = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return &storage.StorerStub{
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromNonce([]byte(common.GetEquivalentProofNonceShardKey(123, core.MetachainShardId)), 0)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work and send to self", func(t *testing.T) {
		t.Parallel()

		providedBuff := []byte("provided buff")
		args := createMockArgEquivalentProofsRequester()
		args.Storage = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (chainStorage.Storer, error) {
				return &storage.StorerStub{
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return providedBuff, nil
					},
				}, nil
			},
		}
		wasSendToConnectedPeerCalled := false
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				wasSendToConnectedPeerCalled = true
				require.Equal(t, string(providedBuff), string(buff))
				return nil
			},
		}
		req, err := NewEquivalentProofsRequester(args)
		require.NoError(t, err)

		err = req.RequestDataFromNonce([]byte(common.GetEquivalentProofNonceShardKey(123, 1)), 0)
		require.NoError(t, err)
		require.True(t, wasSendToConnectedPeerCalled)
	})
}
