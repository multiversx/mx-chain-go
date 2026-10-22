package blockAPI

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/storage"
	dbmock "github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestInternalHeaderExclusionDecodeCount(t *testing.T) {
	for _, shardID := range []uint32{0, core.MetachainShardId} {
		for _, exclusions := range []bool{false, true} {
			for _, format := range []common.ApiOutputFormat{common.ApiOutputFormatJSON, common.ApiOutputFormatProto} {
				args := createMockArgsAPIBlockProc()
				args.SelfShardID = shardID
				args.HistoryRepo = &dbmock.HistoryRepositoryStub{IsEnabledCalled: func() bool { return false }}
				if exclusions {
					var err error
					args.RoundExclusionHandler, err = common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{{StartRound: 100, EndRound: 199}})
					require.NoError(t, err)
				}
				var header data.HeaderHandler = &block.HeaderV3{Round: 50, LastExecutionResult: &block.ExecutionResultInfo{}}
				if shardID == core.MetachainShardId {
					header = &block.MetaBlockV3{Round: 50, LastExecutionResult: &block.MetaExecutionResultInfo{}}
				}
				marshaller := &marshal.GogoProtoMarshalizer{}
				encoded, err := marshaller.Marshal(header)
				require.NoError(t, err)
				decodes := 0
				args.Marshalizer = &marshallerMock.MarshalizerStub{
					MarshalCalled: marshaller.Marshal,
					UnmarshalCalled: func(obj interface{}, buff []byte) error {
						decodes++
						return marshaller.Unmarshal(obj, buff)
					},
				}
				args.Store = &storageMocks.ChainStorerStub{GetCalled: func(dataRetriever.UnitType, []byte) ([]byte, error) { return encoded, nil }}
				proc, err := CreateAPIInternalBlockProcessor(args)
				require.NoError(t, err)
				if shardID == core.MetachainShardId {
					_, err = proc.GetInternalMetaBlockByHash(format, []byte("header"))
				} else {
					_, err = proc.GetInternalShardBlockByHash(format, []byte("header"))
				}
				require.NoError(t, err)
				expectedDecodes := 1
				if !exclusions && format == common.ApiOutputFormatProto {
					expectedDecodes = 0
				}
				require.Equal(t, expectedDecodes, decodes)
			}
		}
	}
}

func TestInternalMiniblockExclusions(t *testing.T) {
	for _, indexed := range []bool{false, true} {
		for _, peer := range []bool{false, true} {
			for _, round := range []uint64{99, 100, 199, 200} {
				args := createMockArgsAPIBlockProc()
				var err error
				args.RoundExclusionHandler, err = common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{{StartRound: 100, EndRound: 199}})
				require.NoError(t, err)
				metadataReads := 0
				args.HistoryRepo = &dbmock.HistoryRepositoryStub{
					IsEnabledCalled: func() bool { return indexed },
					GetMiniblockMetadataByMiniblockHashCalled: func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
						require.True(t, indexed)
						require.False(t, peer)
						metadataReads++
						return &dblookupext.MiniblockMetadata{Round: round}, nil
					},
				}
				mb := &block.MiniBlock{Type: block.TxBlock}
				if peer {
					mb.Type = block.PeerBlock
				}
				encoded, err := args.Marshalizer.Marshal(mb)
				require.NoError(t, err)
				args.Store = &storageMocks.ChainStorerStub{GetStorerCalled: func(dataRetriever.UnitType) (storage.Storer, error) {
					return &storageMocks.StorerStub{GetFromEpochCalled: func([]byte, uint32) ([]byte, error) { return encoded, nil }}, nil
				}}
				proc, err := CreateAPIInternalBlockProcessor(args)
				require.NoError(t, err)
				for _, format := range []common.ApiOutputFormat{common.ApiOutputFormatJSON, common.ApiOutputFormatProto} {
					_, err = proc.GetInternalMiniBlock(format, []byte("mb"), 1)
					if indexed && !peer && args.RoundExclusionHandler.IsRoundExcluded(round) {
						require.ErrorIs(t, err, errBlockNotFound)
					} else {
						require.NoError(t, err)
					}
				}
				if indexed && !peer {
					require.Equal(t, 2, metadataReads)
				} else {
					require.Zero(t, metadataReads)
				}
			}
		}
	}
}

func TestInternalBlockExclusions(t *testing.T) {
	for _, shardID := range []uint32{0, core.MetachainShardId} {
		for _, v3 := range []bool{false, true} {
			for _, round := range []uint64{99, 100, 150, 199, 200, 300, 399, 400} {
				t.Run(fmt.Sprintf("shard=%d/v3=%t/round=%d", shardID, v3, round), func(t *testing.T) {
					args := createMockArgsAPIBlockProc()
					args.SelfShardID = shardID
					args.Marshalizer = &marshal.GogoProtoMarshalizer{}
					var err error
					args.RoundExclusionHandler, err = common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{
						{StartRound: 100, EndRound: 199}, {StartRound: 300, EndRound: 399},
					})
					require.NoError(t, err)
					var header data.HeaderHandler = &block.Header{Round: round}
					if v3 {
						header = &block.HeaderV3{Round: round, LastExecutionResult: &block.ExecutionResultInfo{}}
					}
					if shardID == core.MetachainShardId {
						header = &block.MetaBlock{Round: round}
						if v3 {
							header = &block.MetaBlockV3{Round: round, LastExecutionResult: &block.MetaExecutionResultInfo{}}
						}
					}
					encoded, err := args.Marshalizer.Marshal(header)
					require.NoError(t, err)
					args.Store = &storageMocks.ChainStorerStub{
						GetCalled: func(unit dataRetriever.UnitType, _ []byte) ([]byte, error) {
							if unit == dataRetriever.BlockHeaderUnit || unit == dataRetriever.MetaBlockUnit {
								return encoded, nil
							}
							return []byte("header"), nil
						},
						GetStorerCalled: func(dataRetriever.UnitType) (storage.Storer, error) {
							return &storageMocks.StorerStub{GetFromEpochCalled: func([]byte, uint32) ([]byte, error) { return encoded, nil }}, nil
						},
					}
					proc, err := CreateAPIInternalBlockProcessor(args)
					require.NoError(t, err)
					for _, format := range []common.ApiOutputFormat{common.ApiOutputFormatJSON, common.ApiOutputFormatProto} {
						calls := []func() (interface{}, error){
							func() (interface{}, error) { return proc.GetInternalShardBlockByHash(format, []byte("header")) },
							func() (interface{}, error) { return proc.GetInternalShardBlockByNonce(format, 1) },
							func() (interface{}, error) { return proc.GetInternalShardBlockByRound(format, round) },
						}
						if shardID == core.MetachainShardId {
							calls = []func() (interface{}, error){
								func() (interface{}, error) { return proc.GetInternalMetaBlockByHash(format, []byte("header")) },
								func() (interface{}, error) { return proc.GetInternalMetaBlockByNonce(format, 1) },
								func() (interface{}, error) { return proc.GetInternalMetaBlockByRound(format, round) },
								func() (interface{}, error) { return proc.GetInternalStartOfEpochMetaBlock(format, 1) },
								func() (interface{}, error) { return proc.GetInternalStartOfEpochValidatorsInfo(1) },
							}
						}
						for _, call := range calls {
							_, err = call()
							if args.RoundExclusionHandler.IsRoundExcluded(round) {
								require.ErrorIs(t, err, errBlockNotFound)
							} else {
								require.NoError(t, err)
							}
						}
					}
				})
			}
		}
	}
}
