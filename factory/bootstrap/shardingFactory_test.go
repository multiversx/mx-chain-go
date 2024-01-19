package bootstrap

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	errErd "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	validatorInfoCacherMocks "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

const defaultNumStoredEpochs = 4

func TestCreateShardCoordinator(t *testing.T) {
	t.Parallel()

	t.Run("nil nodes config should error", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(nil, nil, config.PreferencesConfig{}, nil)
		require.Equal(t, errErd.ErrNilGenesisNodesSetupHandler, err)
		require.Empty(t, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("nil pub key should error", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(&testscommon.NodesSetupStub{}, nil, config.PreferencesConfig{}, nil)
		require.Equal(t, errErd.ErrNilPublicKey, err)
		require.Empty(t, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(&testscommon.NodesSetupStub{}, &cryptoMocks.PublicKeyStub{}, config.PreferencesConfig{}, nil)
		require.Equal(t, errErd.ErrNilLogger, err)
		require.Empty(t, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("getShardIdFromNodePubKey fails should error", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(
			&testscommon.NodesSetupStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return nil, expectedErr
				},
			},
			config.PreferencesConfig{},
			&testscommon.LoggerStub{},
		)
		require.Equal(t, expectedErr, err)
		require.Empty(t, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("public key not in genesis - ProcessDestinationShardAsObserver fails should error", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(
			&testscommon.NodesSetupStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return nil, sharding.ErrPublicKeyNotFoundInGenesis // force this error here
				},
			},
			config.PreferencesConfig{
				DestinationShardAsObserver: "", // ProcessDestinationShardAsObserver fails
			},
			&testscommon.LoggerStub{},
		)
		require.NotNil(t, err)
		require.Empty(t, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("public key not in genesis, destination shard disabled - ToByteArray fails should error", func(t *testing.T) {
		t.Parallel()

		counter := 0
		shardC, nodeType, err := CreateShardCoordinator(
			&testscommon.NodesSetupStub{
				GetShardIDForPubKeyCalled: func(pubKey []byte) (uint32, error) {
					return 0, sharding.ErrPublicKeyNotFoundInGenesis // force this error
				},
			},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					counter++
					if counter > 1 {
						return nil, expectedErr
					}
					return []byte("public key"), nil
				},
			},
			config.PreferencesConfig{
				DestinationShardAsObserver: "disabled", // force if branch
			},
			&testscommon.LoggerStub{},
		)
		require.NotNil(t, err)
		require.True(t, errors.Is(err, expectedErr))
		require.Equal(t, core.NodeTypeObserver, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("public key not in genesis, destination shard disabled - should work", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(
			&testscommon.NodesSetupStub{
				GetShardIDForPubKeyCalled: func(pubKey []byte) (uint32, error) {
					return 0, sharding.ErrPublicKeyNotFoundInGenesis // force this error
				},
				NumberOfShardsCalled: func() uint32 {
					return 2
				},
			},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return []byte("public key"), nil
				},
			},
			config.PreferencesConfig{
				DestinationShardAsObserver: "disabled", // for coverage
			},
			&testscommon.LoggerStub{},
		)
		require.Nil(t, err)
		require.Equal(t, core.NodeTypeObserver, nodeType)
		require.False(t, check.IfNil(shardC))
	})
	t.Run("metachain but 0 shards should error", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(
			&testscommon.NodesSetupStub{
				GetShardIDForPubKeyCalled: func(pubKey []byte) (uint32, error) {
					return core.MetachainShardId, nil
				},
				NumberOfShardsCalled: func() uint32 {
					return 0
				},
			},
			&cryptoMocks.PublicKeyStub{},
			config.PreferencesConfig{},
			&testscommon.LoggerStub{},
		)
		require.NotNil(t, err)
		require.Empty(t, nodeType)
		require.True(t, check.IfNil(shardC))
	})
	t.Run("metachain should work", func(t *testing.T) {
		t.Parallel()

		shardC, nodeType, err := CreateShardCoordinator(
			&testscommon.NodesSetupStub{
				GetShardIDForPubKeyCalled: func(pubKey []byte) (uint32, error) {
					return core.MetachainShardId, nil
				},
			},
			&cryptoMocks.PublicKeyStub{},
			config.PreferencesConfig{},
			&testscommon.LoggerStub{},
		)
		require.Nil(t, err)
		require.Equal(t, core.NodeTypeValidator, nodeType)
		require.False(t, check.IfNil(shardC))
	})
}

func TestCreateNodesCoordinator(t *testing.T) {
	t.Parallel()

	t.Run("nil nodes shuffler out closer should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			nil,
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Equal(t, errErd.ErrNilShuffleOutCloser, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("nil nodes config should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			nil,
			config.PreferencesConfig{},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Equal(t, errErd.ErrNilGenesisNodesSetupHandler, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("nil epoch start notifier should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{},
			nil,
			&cryptoMocks.PublicKeyStub{},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Equal(t, errErd.ErrNilEpochStartNotifier, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("nil pub key should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{},
			&mock.EpochStartNotifierStub{},
			nil,
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Equal(t, errErd.ErrNilPublicKey, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("nil bootstrap params should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			nil,
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Equal(t, errErd.ErrNilBootstrapParamsHandler, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("nil chan should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			nil,
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Equal(t, nodesCoordinator.ErrNilNodeStopChannel, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("invalid shard should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.NotNil(t, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("destination shard disabled - ToByteArray fails should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "disabled",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return nil, expectedErr
				},
			},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.True(t, errors.Is(err, expectedErr))
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("ToByteArray fails should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "0",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return nil, expectedErr
				},
			},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.True(t, errors.Is(err, expectedErr))
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("NewShuffledOutTrigger fails should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "0",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return nil, nil // no error but nil pub key to force NewShuffledOutTrigger to fail
				},
			},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.NotNil(t, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("NewIndexHashedNodesCoordinator fails should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "0",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return []byte("public key"), nil
				},
			},
			nil, // force NewIndexHashedNodesCoordinator to fail
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.NotNil(t, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("NewIndexHashedNodesCoordinatorWithRater fails should error", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "0",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return []byte("public key"), nil
				},
			},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			nil, // force NewIndexHashedNodesCoordinatorWithRater to fail
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{
				NodesConfigCalled: func() *nodesCoordinator.NodesCoordinatorRegistry {
					return &nodesCoordinator.NodesCoordinatorRegistry{
						EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
							"0": {
								EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
									"4294967295": {
										{
											PubKey:  []byte("pk1"),
											Chances: 1,
											Index:   0,
										},
									},
								},
								WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
								LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
							},
						},
						CurrentEpoch: 0,
					}
				},
			},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.NotNil(t, err)
		require.True(t, check.IfNil(nodesC))
	})
	t.Run("should work with nodes config", func(t *testing.T) {
		t.Parallel()

		nodesC, err := CreateNodesCoordinator(
			&testscommon.ShuffleOutCloserStub{},
			&testscommon.NodesSetupStub{},
			config.PreferencesConfig{
				DestinationShardAsObserver: "disabled",
			},
			&mock.EpochStartNotifierStub{},
			&cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return []byte("public key"), nil
				},
			},
			&marshallerMock.MarshalizerStub{},
			&testscommon.HasherStub{},
			&testscommon.RaterMock{},
			&storage.StorerStub{},
			&shardingMocks.NodeShufflerMock{},
			0,
			&bootstrapMocks.BootstrapParamsHandlerMock{
				NodesConfigCalled: func() *nodesCoordinator.NodesCoordinatorRegistry {
					return &nodesCoordinator.NodesCoordinatorRegistry{
						EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
							"0": {
								EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
									"4294967295": {
										{
											PubKey:  []byte("pk1"),
											Chances: 1,
											Index:   0,
										},
									},
								},
								WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
								LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
							},
						},
						CurrentEpoch: 0,
					}
				},
			},
			0,
			make(chan endProcess.ArgEndProcess, 1),
			&nodeTypeProviderMock.NodeTypeProviderStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&validatorInfoCacherMocks.ValidatorInfoCacherStub{},
			defaultNumStoredEpochs,
			&storage.StorerStub{},
		)
		require.Nil(t, err)
		require.False(t, check.IfNil(nodesC))
	})
}

func TestCreateNodesShuffleOut(t *testing.T) {
	t.Parallel()

	t.Run("nil nodes config should error", func(t *testing.T) {
		t.Parallel()

		shuffler, err := CreateNodesShuffleOut(nil, config.EpochStartConfig{}, make(chan endProcess.ArgEndProcess, 1))
		require.Equal(t, errErd.ErrNilGenesisNodesSetupHandler, err)
		require.True(t, check.IfNil(shuffler))
	})
	t.Run("invalid MaxShuffledOutRestartThreshold should error", func(t *testing.T) {
		t.Parallel()

		shuffler, err := CreateNodesShuffleOut(
			&testscommon.NodesSetupStub{},
			config.EpochStartConfig{
				MaxShuffledOutRestartThreshold: 5.0,
			},
			make(chan endProcess.ArgEndProcess, 1),
		)
		require.True(t, strings.Contains(err.Error(), "invalid max threshold for shuffled out handler"))
		require.True(t, check.IfNil(shuffler))
	})
	t.Run("invalid MaxShuffledOutRestartThreshold should error", func(t *testing.T) {
		t.Parallel()

		shuffler, err := CreateNodesShuffleOut(
			&testscommon.NodesSetupStub{},
			config.EpochStartConfig{
				MinShuffledOutRestartThreshold: 5.0,
			},
			make(chan endProcess.ArgEndProcess, 1),
		)
		require.True(t, strings.Contains(err.Error(), "invalid min threshold for shuffled out handler"))
		require.True(t, check.IfNil(shuffler))
	})
	t.Run("NewShuffleOutCloser fails should error", func(t *testing.T) {
		t.Parallel()

		shuffler, err := CreateNodesShuffleOut(
			&testscommon.NodesSetupStub{},
			config.EpochStartConfig{},
			nil, // force NewShuffleOutCloser to fail
		)
		require.NotNil(t, err)
		require.True(t, check.IfNil(shuffler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shuffler, err := CreateNodesShuffleOut(
			&testscommon.NodesSetupStub{
				GetRoundDurationCalled: func() uint64 {
					return 4000
				},
			},
			config.EpochStartConfig{
				RoundsPerEpoch:                 200,
				MinShuffledOutRestartThreshold: 0.05,
				MaxShuffledOutRestartThreshold: 0.25,
			},
			make(chan endProcess.ArgEndProcess, 1),
		)
		require.Nil(t, err)
		require.False(t, check.IfNil(shuffler))
	})
}
