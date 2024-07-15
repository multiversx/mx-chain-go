package metachain

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	vics "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/require"
)

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
