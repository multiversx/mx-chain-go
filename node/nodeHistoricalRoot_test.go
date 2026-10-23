package node_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

type boundaryReaderStub struct {
	storage.Storer
	called bool
}

func (s *boundaryReaderStub) GetFromEpochOrNext(key []byte, epoch uint32) ([]byte, error) {
	s.called = true
	return s.Storer.GetFromEpoch(key, epoch)
}

func TestNode_HistoricalV3AccountRoot(t *testing.T) {
	for _, byHash := range []bool{true, false} {
		queryName := "nonce"
		if byHash {
			queryName = "hash"
		}
		for _, scenario := range []string{"historical result", "missing result", "malformed result", "storer unavailable"} {
			t.Run(queryName+"/"+scenario, func(t *testing.T) {
				coreComponents := getDefaultCoreComponents()
				dataComponents := getDefaultDataComponents()
				processComponents := getDefaultProcessComponents()
				const epoch = uint32(7)
				const nonce = uint64(42)
				headerHash := []byte("requested-header")
				rootHash := []byte("requested-execution-root")
				header := &block.HeaderV3{
					Nonce: nonce, Epoch: epoch,
					LastExecutionResult: &block.ExecutionResultInfo{ExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: nonce - 3, RootHash: []byte("earlier-root-must-not-be-used"),
					}},
				}
				headerBytes, err := coreComponents.InternalMarshalizer().Marshal(header)
				require.NoError(t, err)
				chainStore := genericMocks.NewChainStorerMock(epoch + 10)
				require.NoError(t, chainStore.BlockHeaders.PutInEpoch(headerHash, headerBytes, epoch))
				require.NoError(t, chainStore.ShardHdrNonce.Put(
					coreComponents.Uint64ByteSliceConverter().ToByteSlice(nonce), headerHash))
				processComponents.HistoryRepositoryInternal = &dblookupext.HistoryRepositoryStub{
					IsEnabledCalled: func() bool { return true },
					GetEpochByHashCalled: func(hash []byte) (uint32, error) {
						require.Equal(t, headerHash, hash)
						return epoch, nil
					},
				}
				resultBytes, err := coreComponents.InternalMarshalizer().Marshal(&block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: headerHash, HeaderNonce: nonce, HeaderEpoch: epoch, RootHash: rootHash,
					},
				})
				require.NoError(t, err)
				lookupError := errors.New("execution result unavailable")
				reads := 0
				boundaryReader := &boundaryReaderStub{}
				dataComponents.Store = &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unit dataRetriever.UnitType) (storage.Storer, error) {
						if unit != dataRetriever.ExecutionResultsUnit {
							return chainStore.GetStorer(unit)
						}
						if scenario == "storer unavailable" {
							return nil, lookupError
						}
						storer := &storageMocks.StorerStub{
							GetCalled: func([]byte) ([]byte, error) {
								t.Fatal("historical execution results require an epoch-aware lookup")
								return nil, lookupError
							},
							GetFromEpochCalled: func(key []byte, requestedEpoch uint32) ([]byte, error) {
								reads++
								require.Equal(t, headerHash, key)
								require.Equal(t, epoch, requestedEpoch)
								switch scenario {
								case "missing result":
									return nil, lookupError
								case "malformed result":
									return []byte{0xff}, nil
								default:
									return resultBytes, nil
								}
							},
						}
						if byHash {
							boundaryReader.Storer = storer
							return boundaryReader, nil
						}
						return storer, nil
					},
				}
				n, err := node.NewNode(
					node.WithCoreComponents(coreComponents),
					node.WithDataComponents(dataComponents),
					node.WithProcessComponents(processComponents),
				)
				require.NoError(t, err)
				query := api.AccountQueryOptions{BlockNonce: core.OptionalUint64{Value: nonce, HasValue: true}}
				if byHash {
					query = api.AccountQueryOptions{BlockHash: headerHash}
				}
				options, err := n.AddBlockCoordinatesToAccountQueryOptions(query)
				if scenario == "historical result" {
					require.NoError(t, err)
					require.Equal(t, rootHash, options.BlockRootHash)
					require.Equal(t, headerHash, options.BlockHash)
					require.Equal(t, core.OptionalUint64{Value: nonce, HasValue: true}, options.BlockNonce)
					require.Equal(t, core.OptionalUint32{Value: epoch, HasValue: true}, options.HintEpoch)
				} else {
					require.Error(t, err)
					require.Equal(t, api.AccountQueryOptions{}, options)
					if scenario != "malformed result" {
						require.ErrorIs(t, err, lookupError)
					}
				}
				if scenario != "storer unavailable" {
					require.Equal(t, 1, reads)
					require.Equal(t, byHash, boundaryReader.called)
				}
			})
		}
	}
}
