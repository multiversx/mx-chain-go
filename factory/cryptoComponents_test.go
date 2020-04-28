package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoComponentsFactory_NilConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.Config = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilConfiguration, err)
}

func TestNewCryptoComponentsFactory_NiNodesConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.NodesConfig = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilNodesConfig, err)
}

func TestNewCryptoComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.ShardCoordinator = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewCryptoComponentsFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.KeyGen = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilKeyGen, err)
}

func TestNewCryptoComponentsFactory_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.PrivKey = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilPrivateKey, err)
}

func TestNewCryptoComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestCryptoComponentsFactory_CreateShouldErrDueToBadConfig(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.Config = &config.Config{}
	ccf, _ := factory.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.Error(t, err)
	require.Nil(t, cc)
}

func TestCryptoComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, _ := factory.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func getCryptoArgs() factory.CryptoComponentsFactoryArgs {
	return factory.CryptoComponentsFactoryArgs{
		Config: &config.Config{
			Consensus:      config.TypeConfig{Type: "bls"},
			MultisigHasher: config.TypeConfig{Type: "blake2b"},
			Hasher:         config.TypeConfig{Type: "blake2b"},
		},
		NodesConfig:      &mock.NodesSetupStub{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		KeyGen:           &mock.KeyGenMock{},
		PrivKey:          &mock.PrivateKeyMock{},
	}
}
