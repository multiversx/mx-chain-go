package bootstrap_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/mock"
	mock2 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/require"
)

func TestNewEpochStartDataProvider_NilPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.PublicKey = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilPublicKey, err)
}

func TestNewEpochStartDataProvider_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Messenger = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilMessenger, err)
}

func TestNewEpochStartDataProvider_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Marshalizer = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilMarshalizer, err)
}

func TestNewEpochStartDataProvider_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Hasher = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilHasher, err)
}

func TestNewEpochStartDataProvider_NilNodesConfigProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.NodesConfigProvider = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilNodesConfigProvider, err)
}

func TestNewEpochStartDataProvider_NilMetablockInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MetaBlockInterceptor = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilMetaBlockInterceptor, err)
}

func TestNewEpochStartDataProvider_NilShardHeaderInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.ShardHeaderInterceptor = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilShardHeaderInterceptor, err)
}

func TestNewEpochStartDataProvider_NilMetaBlockInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MetaBlockInterceptor = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilMetaBlockInterceptor, err)
}

func TestNewEpochStartDataProvider_NilMiniBlockInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.MiniBlockInterceptor = nil
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, epStart)
	require.Equal(t, bootstrap.ErrNilMiniBlockInterceptor, err)
}

func TestNewEpochStartDataProvider_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getArguments()
	epStart, err := bootstrap.NewEpochStartDataProvider(args)

	require.Nil(t, err)
	require.False(t, check.IfNil(epStart))
}

func TestEpochStartDataProvider_Bootstrap_TopicCreationFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error while creating topic")
	args := getArguments()
	args.Messenger = &mock.MessengerStub{
		CreateTopicCalled: func(_ string, _ bool) error {
			return expectedErr
		},
	}
	epStart, _ := bootstrap.NewEpochStartDataProvider(args)

	res, err := epStart.Bootstrap()

	require.Nil(t, res)
	require.Equal(t, expectedErr, err)
}

func getArguments() bootstrap.ArgsEpochStartDataProvider {
	return bootstrap.ArgsEpochStartDataProvider{
		PublicKey:                      &mock.PublicKeyMock{},
		Messenger:                      &mock.MessengerStub{},
		Marshalizer:                    &mock2.MarshalizerMock{},
		Hasher:                         mock2.HasherMock{},
		NodesConfigProvider:            &mock.NodesConfigProviderStub{},
		EpochStartMetaBlockInterceptor: &mock.EpochStartMetaBlockInterceptorStub{},
		MetaBlockInterceptor:           &mock.MetaBlockInterceptorStub{},
		ShardHeaderInterceptor:         &mock.ShardHeaderInterceptorStub{},
		MiniBlockInterceptor:           &mock.MiniBlockInterceptorStub{},
	}
}
