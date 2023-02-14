package factory_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/require"
)

func createMockNotifierFactoryArgs() *factory.EventNotifierFactoryArgs {
	return &factory.EventNotifierFactoryArgs{
		Enabled:           true,
		UseAuthorization:  true,
		ProxyUrl:          "http://localhost:5000",
		Username:          "",
		Password:          "",
		RequestTimeoutSec: 1,
		Marshaller:        &testscommon.MarshalizerMock{},
		Hasher:            &hashingMocks.HasherMock{},
		PubKeyConverter:   &testscommon.PubkeyConverterMock{},
	}
}

func TestCreateEventNotifier(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := createMockNotifierFactoryArgs()
		args.Marshaller = nil

		en, err := factory.CreateEventNotifier(args)
		require.Nil(t, en)
		require.Equal(t, core.ErrNilMarshalizer, err)
	})

	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		args := createMockNotifierFactoryArgs()
		args.Hasher = nil

		en, err := factory.CreateEventNotifier(args)
		require.Nil(t, en)
		require.Equal(t, core.ErrNilHasher, err)
	})

	t.Run("nil pub key converter", func(t *testing.T) {
		t.Parallel()

		args := createMockNotifierFactoryArgs()
		args.PubKeyConverter = nil

		en, err := factory.CreateEventNotifier(args)
		require.Nil(t, en)
		require.Equal(t, outport.ErrNilPubKeyConverter, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		en, err := factory.CreateEventNotifier(createMockNotifierFactoryArgs())
		require.Nil(t, err)
		require.NotNil(t, en)
	})
}
