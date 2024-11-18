package metachain

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	vics "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
)

func TestNewSovereignValidatorInfoCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil input, should return error", func(t *testing.T) {
		svic, err := NewSovereignValidatorInfoCreator(nil)
		require.Equal(t, process.ErrNilEpochStartValidatorInfoCreator, err)
		require.Nil(t, svic)
	})
	t.Run("should work", func(t *testing.T) {
		svic, err := NewSovereignValidatorInfoCreator(&validatorInfoCreator{})
		require.Nil(t, err)
		require.False(t, svic.IsInterfaceNil())
	})
}

func TestSovereignValidatorInfoCreator_CreateMarshalledData(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	marshalledSVI1, _ := arguments.Marshalizer.Marshal(svi1)

	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
	marshalledSVI2, _ := arguments.Marshalizer.Marshal(svi2)

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	shardDataCacher := testscommon.NewShardedDataCacheNotifierMock()
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vics.ValidatorInfoCacherStub{
				GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
					if bytes.Equal(validatorInfoHash, txHash1) {
						return svi1, nil
					}
					if bytes.Equal(validatorInfoHash, txHash2) {
						return svi2, nil
					}

					require.Fail(t, "should not call get for another validator")
					return nil, nil
				},
			}
		},
		ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return shardDataCacher
		},
	}
	vic, _ := NewValidatorInfoCreator(arguments)
	svic, _ := NewSovereignValidatorInfoCreator(vic)

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   core.SovereignChainShardId,
				ReceiverShardID: core.SovereignChainShardId,
				Type:            block.PeerBlock,
				TxHashes: [][]byte{
					txHash1,
					txHash2,
				},
			},
			{
				SenderShardID:   core.SovereignChainShardId,
				ReceiverShardID: core.SovereignChainShardId,
				Type:            block.StateBlock,
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		},
	}
	marshalledData := svic.CreateMarshalledData(body)
	require.Equal(t, 1, len(marshalledData))
	require.Equal(t, 2, len(marshalledData[common.ValidatorInfoTopic]))
	require.Equal(t, marshalledSVI1, marshalledData[common.ValidatorInfoTopic][0])
	require.Equal(t, marshalledSVI2, marshalledData[common.ValidatorInfoTopic][1])

	cachedData1, found := shardDataCacher.SearchFirstData(txHash1)
	require.True(t, found)
	require.Equal(t, svi1, cachedData1)

	cachedData2, found := shardDataCacher.SearchFirstData(txHash2)
	require.True(t, found)
	require.Equal(t, svi2, cachedData2)
}

func TestSovereignValidatorInfoCreator_CreateMarshalledDataErrorCases(t *testing.T) {

	t.Run("CreateMarshalledData should return nil body is nil", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		vic, _ := NewValidatorInfoCreator(arguments)
		svic, _ := NewSovereignValidatorInfoCreator(vic)

		marshalledData := svic.CreateMarshalledData(nil)
		require.Nil(t, marshalledData)
	})

	t.Run("CreateMarshalledData should return empty slice when tx hash does not exist in validator info cacher", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		localErr := errors.New("local error")
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{
					GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
						return nil, localErr
					},
				}
			},
		}
		wasMarshallCalled := false
		arguments.Marshalizer = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				wasMarshallCalled = true
				return nil, nil
			},
		}

		vic, _ := NewValidatorInfoCreator(arguments)
		svic, _ := NewSovereignValidatorInfoCreator(vic)

		body := createMockBlockBody(core.SovereignChainShardId, core.SovereignChainShardId, block.PeerBlock)
		marshalledData := svic.CreateMarshalledData(body)
		require.Empty(t, marshalledData)
		require.False(t, wasMarshallCalled)
	})

	t.Run("CreateMarshalledData should return empty slice when marhsall fails", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{
					GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
						return &state.ShardValidatorInfo{}, nil
					},
				}
			},
		}
		localErr := errors.New("local error")

		arguments.Marshalizer = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {

				return nil, localErr
			},
		}
		shardDataCacher := testscommon.NewShardedDataCacheNotifierMock()
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return shardDataCacher
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{}
			},
		}

		vic, _ := NewValidatorInfoCreator(arguments)
		svic, _ := NewSovereignValidatorInfoCreator(vic)

		body := createMockBlockBody(core.SovereignChainShardId, core.SovereignChainShardId, block.PeerBlock)
		marshalledData := svic.CreateMarshalledData(body)
		require.Empty(t, marshalledData)
		require.Empty(t, shardDataCacher.Keys())
	})
}

func TestSovereignValidatorInfoCreator_CreateValidatorInfoMiniBlocks(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.ShardCoordinator = sharding.NewSovereignShardCoordinator()

	v1 := &state.ValidatorInfo{
		ShardId:   core.SovereignChainShardId,
		PublicKey: []byte("pubKey1"),
	}
	v2 := &state.ValidatorInfo{
		ShardId:   core.SovereignChainShardId,
		PublicKey: []byte("pubKey2"),
	}

	v1ShardData := createShardValidatorInfo(v1)
	v2ShardData := createShardValidatorInfo(v2)

	txHash1, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, v1ShardData)
	txHash2, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, v2ShardData)

	addDataCt := 0
	valInfoCache := &vics.ValidatorInfoCacherStub{
		AddValidatorInfoCalled: func(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
			switch addDataCt {
			case 0:
				require.Equal(t, txHash1, validatorInfoHash)
				require.Equal(t, v1ShardData, validatorInfo)
			case 1:
				require.Equal(t, txHash2, validatorInfoHash)
				require.Equal(t, v2ShardData, validatorInfo)
			}

			addDataCt++
		},
	}

	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return valInfoCache
		},
	}

	vic, _ := NewValidatorInfoCreator(arguments)
	svic, _ := NewSovereignValidatorInfoCreator(vic)

	valMap := state.NewShardValidatorsInfoMap()
	_ = valMap.Add(v1)
	_ = valMap.Add(v2)

	mb, err := svic.CreateValidatorInfoMiniBlocks(valMap)
	require.Nil(t, err)
	require.NotNil(t, mb)
	require.Equal(t, block.MiniBlockSlice{
		{
			TxHashes:        [][]byte{txHash1, txHash2},
			ReceiverShardID: core.SovereignChainShardId,
			SenderShardID:   core.SovereignChainShardId,
			Type:            block.PeerBlock,
		},
	}, mb)
	require.Equal(t, 2, addDataCt)
}
