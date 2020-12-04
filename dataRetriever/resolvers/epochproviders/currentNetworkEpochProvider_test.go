package epochproviders_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/require"
)

func TestNewCurrentNetworkEpochProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		argsFunc     func() epochproviders.ArgsCurrentNetworkProvider
		expectedErr  error
		shouldNotErr bool
	}{
		{
			argsFunc: func() epochproviders.ArgsCurrentNetworkProvider {
				args := getArgs()
				args.Messenger = nil
				return args
			},
			expectedErr: epochproviders.ErrNilMessenger,
		},
		{
			argsFunc: func() epochproviders.ArgsCurrentNetworkProvider {
				args := getArgs()
				args.RequestHandler = nil
				return args
			},
			expectedErr: epochproviders.ErrNilRequestHandler,
		},
		{
			argsFunc: func() epochproviders.ArgsCurrentNetworkProvider {
				args := getArgs()
				args.EpochStartMetaBlockInterceptor = nil
				return args
			},
			expectedErr: epochproviders.ErrNilEpochStartMetaBlockInterceptor,
		},
		{
			argsFunc: func() epochproviders.ArgsCurrentNetworkProvider {
				args := getArgs()
				return args
			},
			expectedErr:  nil,
			shouldNotErr: true,
		},
	}

	for _, tt := range tests {
		cnep, err := epochproviders.NewCurrentNetworkEpochProvider(tt.argsFunc())
		if !tt.shouldNotErr {
			require.True(t, errors.Is(err, tt.expectedErr))
			require.True(t, check.IfNil(cnep))
		} else {
			require.NotNil(t, cnep)
			require.NoError(t, err)
		}
	}
}

func TestCurrentNetworkEpochProvider_CurrentEpoch(t *testing.T) {
	t.Parallel()

	cnep, _ := epochproviders.NewCurrentNetworkEpochProvider(getArgs())

	expEpoch := uint32(37)
	cnep.SetNetworkEpochAtBootstrap(expEpoch)
	require.Equal(t, cnep.CurrentEpoch(), expEpoch)
}

func TestCurrentNetworkEpochProvider_EpochIsActiveInNetwork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.NumActivePersisters = 3
	cnep, _ := epochproviders.NewCurrentNetworkEpochProvider(args)

	tests := []struct {
		networkEpoch uint32
		nodeEpoch    uint32
		output       bool
		description  string
	}{
		{
			networkEpoch: 5,
			nodeEpoch:    0,
			output:       false,
			description:  "0 not in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    1,
			output:       false,
			description:  "1 not in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    2,
			output:       false,
			description:  "2 not in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    6,
			output:       false,
			description:  "6 not in (3, 5]",
		},
	}

	for _, tt := range tests {
		cnep.SetNetworkEpochAtBootstrap(tt.networkEpoch)
		testOk := tt.output == cnep.EpochIsActiveInNetwork(tt.nodeEpoch)
		if !testOk {
			require.Failf(t, "%s for epoch param %d and network epoch %d",
				tt.description,
				tt.nodeEpoch,
				tt.networkEpoch,
			)
		}
	}
}

func getArgs() epochproviders.ArgsCurrentNetworkProvider {
	return epochproviders.ArgsCurrentNetworkProvider{
		RequestHandler: &mock.RequestHandlerStub{},
		Messenger: &mock.MessengerStub{
			RegisterMessageProcessorCalled: func(_ string, _ string, _ p2p.MessageProcessor) error {
				return nil
			},
			UnregisterMessageProcessorCalled: func(_ string, _ string) error {
				return nil
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"0", "1"}
			},
		},
		EpochStartMetaBlockInterceptor: &mock.InterceptorStub{},
		NumActivePersisters:            2,
		DurationBetweenMetablockChecks: time.Minute,
	}
}
