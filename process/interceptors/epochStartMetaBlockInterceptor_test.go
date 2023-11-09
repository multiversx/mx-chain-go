package interceptors_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func TestNewEpochStartMetaBlockInterceptor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description  string
		argsFunc     func() interceptors.ArgsEpochStartMetaBlockInterceptor
		exptectedErr error
		shouldNotErr bool
	}{
		{
			description: "nil marshalizer",
			argsFunc: func() interceptors.ArgsEpochStartMetaBlockInterceptor {
				args := getArgsEpochStartMetaBlockInterceptor()
				args.Marshalizer = nil
				return args
			},
			exptectedErr: process.ErrNilMarshalizer,
		},
		{
			description: "nil hasher",
			argsFunc: func() interceptors.ArgsEpochStartMetaBlockInterceptor {
				args := getArgsEpochStartMetaBlockInterceptor()
				args.Hasher = nil
				return args
			},
			exptectedErr: process.ErrNilHasher,
		},
		{
			description: "nil num connected peers provider",
			argsFunc: func() interceptors.ArgsEpochStartMetaBlockInterceptor {
				args := getArgsEpochStartMetaBlockInterceptor()
				args.NumConnectedPeersProvider = nil
				return args
			},
			exptectedErr: process.ErrNilNumConnectedPeersProvider,
		},
		{
			description: "consensus percentage lower than 0",
			argsFunc: func() interceptors.ArgsEpochStartMetaBlockInterceptor {
				args := getArgsEpochStartMetaBlockInterceptor()
				args.ConsensusPercentage = -10
				return args
			},
			exptectedErr: process.ErrInvalidEpochStartMetaBlockConsensusPercentage,
		},
		{
			description: "consensus percentage higher than 100",
			argsFunc: func() interceptors.ArgsEpochStartMetaBlockInterceptor {
				args := getArgsEpochStartMetaBlockInterceptor()
				args.ConsensusPercentage = 110
				return args
			},
			exptectedErr: process.ErrInvalidEpochStartMetaBlockConsensusPercentage,
		},
		{
			description: "all constructor parameters ok",
			argsFunc: func() interceptors.ArgsEpochStartMetaBlockInterceptor {
				args := getArgsEpochStartMetaBlockInterceptor()
				return args
			},
			exptectedErr: nil,
			shouldNotErr: true,
		},
	}

	for _, tt := range tests {
		esmbi, err := interceptors.NewEpochStartMetaBlockInterceptor(tt.argsFunc())
		if tt.shouldNotErr {
			require.NoError(t, err)
			require.False(t, check.IfNil(esmbi))
			continue
		}

		require.True(t, errors.Is(err, tt.exptectedErr))
		require.True(t, check.IfNil(esmbi))
	}
}

func TestEpochStartMetaBlockInterceptor_ProcessReceivedMessageUnmarshalError(t *testing.T) {
	t.Parallel()

	esmbi, _ := interceptors.NewEpochStartMetaBlockInterceptor(getArgsEpochStartMetaBlockInterceptor())
	require.NotNil(t, esmbi)

	message := &p2pmocks.P2PMessageMock{DataField: []byte("wrong meta block  bytes")}
	err := esmbi.ProcessReceivedMessage(message, "", &p2pmocks.MessengerStub{})
	require.Error(t, err)
}

func TestEpochStartMetaBlockInterceptor_EntireFlowShouldWorkAndSetTheEpoch(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(37)
	wasCalled := false
	args := getArgsEpochStartMetaBlockInterceptor()
	handlerFunc := func(topic string, hash []byte, data interface{}) {
		mbEpoch := data.(*block.MetaBlock).Epoch
		require.Equal(t, expectedEpoch, mbEpoch)
		wasCalled = true
	}
	args.NumConnectedPeersProvider = &p2pmocks.MessengerStub{
		ConnectedPeersCalled: func() []core.PeerID {
			return make([]core.PeerID, 6) // 6 connected peers
		},
	}
	args.ConsensusPercentage = 50 // 50% , so at least 3/6 have to send the same meta block

	esmbi, _ := interceptors.NewEpochStartMetaBlockInterceptor(args)
	require.NotNil(t, esmbi)
	esmbi.RegisterHandler(handlerFunc)

	metaBlock := &block.MetaBlock{
		Epoch: expectedEpoch,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
				},
				{
					ShardID: 1,
				},
			},
		},
	}
	metaBlockBytes, _ := args.Marshalizer.Marshal(metaBlock)

	wrongMetaBlock := &block.MetaBlock{Epoch: 0}
	wrongMetaBlockBytes, _ := args.Marshalizer.Marshal(wrongMetaBlock)

	err := esmbi.ProcessReceivedMessage(&p2pmocks.P2PMessageMock{DataField: metaBlockBytes}, "peer0", &p2pmocks.MessengerStub{})
	require.NoError(t, err)
	require.False(t, wasCalled)

	_ = esmbi.ProcessReceivedMessage(&p2pmocks.P2PMessageMock{DataField: metaBlockBytes}, "peer1", &p2pmocks.MessengerStub{})
	require.False(t, wasCalled)

	// send again from peer1 => should not be taken into account
	_ = esmbi.ProcessReceivedMessage(&p2pmocks.P2PMessageMock{DataField: metaBlockBytes}, "peer1", &p2pmocks.MessengerStub{})
	require.False(t, wasCalled)

	// send another meta block
	_ = esmbi.ProcessReceivedMessage(&p2pmocks.P2PMessageMock{DataField: wrongMetaBlockBytes}, "peer2", &p2pmocks.MessengerStub{})
	require.False(t, wasCalled)

	// send the last needed metablock from a new peer => should fetch the epoch
	_ = esmbi.ProcessReceivedMessage(&p2pmocks.P2PMessageMock{DataField: metaBlockBytes}, "peer3", &p2pmocks.MessengerStub{})
	require.True(t, wasCalled)

}

func getArgsEpochStartMetaBlockInterceptor() interceptors.ArgsEpochStartMetaBlockInterceptor {
	return interceptors.ArgsEpochStartMetaBlockInterceptor{
		Marshalizer:               &mock.MarshalizerMock{},
		Hasher:                    &hashingMocks.HasherMock{},
		NumConnectedPeersProvider: &p2pmocks.MessengerStub{},
		ConsensusPercentage:       50,
	}
}
