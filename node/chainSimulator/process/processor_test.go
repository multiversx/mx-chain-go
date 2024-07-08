package process_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	mockConsensus "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainSimulator"
	testsConsensus "github.com/multiversx/mx-chain-go/testscommon/consensus"
	testsFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func TestNewBlocksCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil node handler should error", func(t *testing.T) {
		t.Parallel()

		creator, err := chainSimulatorProcess.NewBlocksCreator(nil)
		require.Equal(t, chainSimulatorProcess.ErrNilNodeHandler, err)
		require.Nil(t, creator)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		creator, err := chainSimulatorProcess.NewBlocksCreator(&chainSimulator.NodeHandlerMock{})
		require.NoError(t, err)
		require.NotNil(t, creator)
	})
}

func TestBlocksCreator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	creator, _ := chainSimulatorProcess.NewBlocksCreator(nil)
	require.True(t, creator.IsInterfaceNil())

	creator, _ = chainSimulatorProcess.NewBlocksCreator(&chainSimulator.NodeHandlerMock{})
	require.False(t, creator.IsInterfaceNil())
}

func TestBlocksCreator_IncrementRound(t *testing.T) {
	t.Parallel()

	wasIncrementIndexCalled := false
	wasSetUInt64ValueCalled := false
	nodeHandler := &chainSimulator.NodeHandlerMock{
		GetCoreComponentsCalled: func() factory.CoreComponentsHolder {
			return &testsFactory.CoreComponentsHolderStub{
				RoundHandlerCalled: func() consensus.RoundHandler {
					return &testscommon.RoundHandlerMock{
						IncrementIndexCalled: func() {
							wasIncrementIndexCalled = true
						},
					}
				},
			}
		},
		GetStatusCoreComponentsCalled: func() factory.StatusCoreComponentsHolder {
			return &testsFactory.StatusCoreComponentsStub{
				AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{
					SetUInt64ValueHandler: func(key string, value uint64) {
						wasSetUInt64ValueCalled = true
						require.Equal(t, common.MetricCurrentRound, key)
					},
				},
			}
		},
	}
	creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
	require.NoError(t, err)

	creator.IncrementRound()
	require.True(t, wasIncrementIndexCalled)
	require.True(t, wasSetUInt64ValueCalled)
}

func TestBlocksCreator_CreateNewBlock(t *testing.T) {
	t.Parallel()

	t.Run("CreateNewHeader failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return nil, expectedErr
			},
		}
		nodeHandler := getNodeHandler()
		nodeHandler.GetProcessComponentsCalled = func() factory.ProcessComponentsHolder {
			return &mock.ProcessComponentsStub{
				BlockProcess: blockProcess,
			}
		}
		nodeHandler.GetChainHandlerCalled = func() data.ChainHandler {
			return &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.HeaderV2{} // coverage for getPreviousHeaderData
				},
			}
		}

		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("SetShardID failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetShardIDCalled: func(shardId uint32) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("SetPrevHash failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetPrevHashCalled: func(hash []byte) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("SetPrevRandSeed failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetPrevRandSeedCalled: func(seed []byte) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("SetPubKeysBitmap failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetPubKeysBitmapCalled: func(bitmap []byte) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("SetChainID failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetChainIDCalled: func(chainID []byte) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("SetTimeStamp failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetTimeStampCalled: func(timestamp uint64) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("ComputeConsensusGroup failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		nodeHandler.GetProcessComponentsCalled = func() factory.ProcessComponentsHolder {
			return &mock.ProcessComponentsStub{
				BlockProcess: &testscommon.BlockProcessorStub{
					CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
						return &testscommon.HeaderHandlerStub{}, nil
					},
				},
				NodesCoord: &shardingMocks.NodesCoordinatorStub{
					ComputeConsensusGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
						return nil, expectedErr
					},
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("key not managed by the current node should return nil", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		nodeHandler.GetCryptoComponentsCalled = func() factory.CryptoComponentsHolder {
			return &mock.CryptoComponentsStub{
				KeysHandlerField: &testscommon.KeysHandlerStub{
					IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
						return false
					},
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.NoError(t, err)
	})
	t.Run("CreateSignatureForPublicKey failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		kh := nodeHandler.GetCryptoComponents().KeysHandler()
		nodeHandler.GetCryptoComponentsCalled = func() factory.CryptoComponentsHolder {
			return &mock.CryptoComponentsStub{
				KeysHandlerField: kh,
				SigHandler: &testsConsensus.SigningHandlerStub{
					CreateSignatureForPublicKeyCalled: func(message []byte, publicKeyBytes []byte) ([]byte, error) {
						return nil, expectedErr
					},
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("SetRandSeed failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{
					SetRandSeedCalled: func(seed []byte) error {
						return expectedErr
					},
				}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("CreateBlock failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return nil, nil, expectedErr
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("setHeaderSignatures.Marshal failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		rh := nodeHandler.GetCoreComponents().RoundHandler()
		nodeHandler.GetCoreComponentsCalled = func() factory.CoreComponentsHolder {
			return &testsFactory.CoreComponentsHolderStub{
				RoundHandlerCalled: func() consensus.RoundHandler {
					return rh
				},
				InternalMarshalizerCalled: func() marshal.Marshalizer {
					return &testscommon.MarshallerStub{
						MarshalCalled: func(obj interface{}) ([]byte, error) {
							return nil, expectedErr
						},
					}
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("setHeaderSignatures.Reset failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		kh := nodeHandler.GetCryptoComponents().KeysHandler()
		nodeHandler.GetCryptoComponentsCalled = func() factory.CryptoComponentsHolder {
			return &mock.CryptoComponentsStub{
				KeysHandlerField: kh,
				SigHandler: &testsConsensus.SigningHandlerStub{
					ResetCalled: func(pubKeys []string) error {
						return expectedErr
					},
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("setHeaderSignatures.CreateSignatureShareForPublicKey failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		kh := nodeHandler.GetCryptoComponents().KeysHandler()
		nodeHandler.GetCryptoComponentsCalled = func() factory.CryptoComponentsHolder {
			return &mock.CryptoComponentsStub{
				KeysHandlerField: kh,
				SigHandler: &testsConsensus.SigningHandlerStub{
					CreateSignatureShareForPublicKeyCalled: func(message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
						return nil, expectedErr
					},
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("setHeaderSignatures.AggregateSigs failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		kh := nodeHandler.GetCryptoComponents().KeysHandler()
		nodeHandler.GetCryptoComponentsCalled = func() factory.CryptoComponentsHolder {
			return &mock.CryptoComponentsStub{
				KeysHandlerField: kh,
				SigHandler: &testsConsensus.SigningHandlerStub{
					AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
						return nil, expectedErr
					},
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("setHeaderSignatures.SetSignature failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return &testscommon.HeaderHandlerStub{
					CloneCalled: func() data.HeaderHandler {
						return &testscommon.HeaderHandlerStub{}
					},
					SetSignatureCalled: func(signature []byte) error {
						return expectedErr
					},
				}, &block.Body{}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("createLeaderSignature.SetLeaderSignature failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return &testscommon.HeaderHandlerStub{
					CloneCalled: func() data.HeaderHandler {
						return &testscommon.HeaderHandlerStub{
							SetLeaderSignatureCalled: func(signature []byte) error {
								return expectedErr
							},
						}
					},
				}, &block.Body{}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("createLeaderSignature.SetLeaderSignature failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return &testscommon.HeaderHandlerStub{
					CloneCalled: func() data.HeaderHandler {
						return &testscommon.HeaderHandlerStub{
							SetLeaderSignatureCalled: func(signature []byte) error {
								return expectedErr
							},
						}
					},
				}, &block.Body{}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("setHeaderSignatures.SetLeaderSignature failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return &testscommon.HeaderHandlerStub{
					CloneCalled: func() data.HeaderHandler {
						return &testscommon.HeaderHandlerStub{}
					},
					SetLeaderSignatureCalled: func(signature []byte) error {
						return expectedErr
					},
				}, &block.Body{}, nil
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("CommitBlock failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return &testscommon.HeaderHandlerStub{
					CloneCalled: func() data.HeaderHandler {
						return &testscommon.HeaderHandlerStub{}
					},
				}, &block.Body{}, nil
			},
			CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
				return expectedErr
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("MarshalizedDataToBroadcast failure should error", func(t *testing.T) {
		t.Parallel()

		blockProcess := &testscommon.BlockProcessorStub{
			CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
				return &testscommon.HeaderHandlerStub{}, nil
			},
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				return &testscommon.HeaderHandlerStub{
					CloneCalled: func() data.HeaderHandler {
						return &testscommon.HeaderHandlerStub{}
					},
				}, &block.Body{}, nil
			},
			MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
				return nil, nil, expectedErr
			},
		}
		testCreateNewBlock(t, blockProcess, expectedErr)
	})
	t.Run("BroadcastHeader failure should error", func(t *testing.T) {
		t.Parallel()

		nodeHandler := getNodeHandler()
		nodeHandler.GetBroadcastMessengerCalled = func() consensus.BroadcastMessenger {
			return &mockConsensus.BroadcastMessengerMock{
				BroadcastHeaderCalled: func(handler data.HeaderHandler, bytes []byte) error {
					return expectedErr
				},
			}
		}
		creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		creator, err := chainSimulatorProcess.NewBlocksCreator(getNodeHandler())
		require.NoError(t, err)

		err = creator.CreateNewBlock()
		require.NoError(t, err)
	})
}

func testCreateNewBlock(t *testing.T, blockProcess process.BlockProcessor, expectedErr error) {
	nodeHandler := getNodeHandler()
	nc := nodeHandler.GetProcessComponents().NodesCoordinator()
	nodeHandler.GetProcessComponentsCalled = func() factory.ProcessComponentsHolder {
		return &mock.ProcessComponentsStub{
			BlockProcess: blockProcess,
			NodesCoord:   nc,
		}
	}
	creator, err := chainSimulatorProcess.NewBlocksCreator(nodeHandler)
	require.NoError(t, err)

	err = creator.CreateNewBlock()
	require.Equal(t, expectedErr, err)
}

func getNodeHandler() *chainSimulator.NodeHandlerMock {
	return &chainSimulator.NodeHandlerMock{
		GetCoreComponentsCalled: func() factory.CoreComponentsHolder {
			return &testsFactory.CoreComponentsHolderStub{
				RoundHandlerCalled: func() consensus.RoundHandler {
					return &testscommon.RoundHandlerMock{
						TimeStampCalled: func() time.Time {
							return time.Now()
						},
					}
				},
				InternalMarshalizerCalled: func() marshal.Marshalizer {
					return &testscommon.MarshallerStub{}
				},
				HasherCalled: func() hashing.Hasher {
					return &testscommon.HasherStub{
						ComputeCalled: func(s string) []byte {
							return []byte("hash")
						},
					}
				},
			}
		},
		GetProcessComponentsCalled: func() factory.ProcessComponentsHolder {
			return &mock.ProcessComponentsStub{
				BlockProcess: &testscommon.BlockProcessorStub{
					CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
						return &testscommon.HeaderHandlerStub{}, nil
					},
					CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
						haveTime() // coverage only
						return &testscommon.HeaderHandlerStub{
							CloneCalled: func() data.HeaderHandler {
								return &testscommon.HeaderHandlerStub{}
							},
						}, &block.Body{}, nil
					},
					MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
						return make(map[uint32][]byte), make(map[string][][]byte), nil
					},
				},
				NodesCoord: &shardingMocks.NodesCoordinatorStub{
					ComputeConsensusGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
						return []nodesCoordinator.Validator{
							shardingMocks.NewValidatorMock([]byte("A"), 1, 1),
						}, nil
					},
				},
			}
		},
		GetChainHandlerCalled: func() data.ChainHandler {
			return &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return &block.HeaderV2{}
				},
			}
		},
		GetShardCoordinatorCalled: func() sharding.Coordinator {
			return &testscommon.ShardsCoordinatorMock{}
		},
		GetCryptoComponentsCalled: func() factory.CryptoComponentsHolder {
			return &mock.CryptoComponentsStub{
				KeysHandlerField: &testscommon.KeysHandlerStub{
					IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
						return true
					},
				},
				SigHandler: &testsConsensus.SigningHandlerStub{},
			}
		},
		GetBroadcastMessengerCalled: func() consensus.BroadcastMessenger {
			return &mockConsensus.BroadcastMessengerMock{}
		},
	}
}
