package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoComponentsFactory_NiNodesConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.NodesConfig = nil
	ccf, err := factory.NewCryptoComponentsFactory(args, false)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilNodesConfig, err)
}

func TestNewCryptoComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.ShardCoordinator = nil
	ccf, err := factory.NewCryptoComponentsFactory(args, false)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewCryptoComponentsFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.KeyGen = nil
	ccf, err := factory.NewCryptoComponentsFactory(args, false)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilKeyGen, err)
}

func TestNewCryptoComponentsFactory_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.PrivKey = nil
	ccf, err := factory.NewCryptoComponentsFactory(args, false)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilPrivateKey, err)
}

func TestNewCryptoComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, err := factory.NewCryptoComponentsFactory(args, false)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestNewCryptoComponentsFactory_InImportDBModeShouldWork(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, err := factory.NewCryptoComponentsFactory(args, true)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestCryptoComponentsFactory_CreateShouldErrDueToBadConfig(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.Config = config.Config{}
	ccf, _ := factory.NewCryptoComponentsFactory(args, false)

	cc, err := ccf.Create()
	require.Error(t, err)
	require.Nil(t, cc)
}

func TestCryptoComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, _ := factory.NewCryptoComponentsFactory(args, false)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func getCryptoArgs() factory.CryptoComponentsFactoryArgs {
	return factory.CryptoComponentsFactoryArgs{
		Config: config.Config{
			Hasher:         config.TypeConfig{Type: "blake2b"},
			MultisigHasher: config.TypeConfig{Type: "blake2b"},
			PublicKeyPIDSignature: config.CacheConfig{
				Capacity: 1000,
				Type:     "LRU",
			},
			Consensus: config.TypeConfig{Type: "bls"},
		},
		NodesConfig:      &mock.NodesSetupStub{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		KeyGen:           &mock.KeyGenMock{},
		PrivKey:          &mock.PrivateKeyMock{},
	}
}
