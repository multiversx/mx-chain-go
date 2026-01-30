package fees_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/fees"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestGetFeesInEpoch(t *testing.T) {
	t.Parallel()

	t.Run("nil header should return 0", func(t *testing.T) {
		t.Parallel()

		result := fees.GetAccumulatedFeesInEpoch(
			nil,
			&pool.HeadersPoolStub{},
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)

		result = fees.GetDeveloperFeesInEpoch(
			nil,
			&pool.HeadersPoolStub{},
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)
	})

	t.Run("nil headers pool should return 0", func(t *testing.T) {
		t.Parallel()

		result := fees.GetAccumulatedFeesInEpoch(
			&block.MetaBlock{},
			nil,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)
	})

	t.Run("nil marshaller should return 0", func(t *testing.T) {
		t.Parallel()

		result := fees.GetAccumulatedFeesInEpoch(
			&block.MetaBlock{},
			&pool.HeadersPoolStub{},
			nil,
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)
	})

	t.Run("nil storage service should return 0", func(t *testing.T) {
		t.Parallel()

		result := fees.GetAccumulatedFeesInEpoch(
			&block.MetaBlock{},
			&pool.HeadersPoolStub{},
			&marshallerMock.MarshalizerStub{},
			nil,
		)
		require.Equal(t, big.NewInt(0), result)
	})

	t.Run("non-v3 meta header should return from header", func(t *testing.T) {
		t.Parallel()

		expectedAccumulatedFees := big.NewInt(3000)
		expectedDevFees := big.NewInt(5000)
		epochStartHeader := &block.MetaBlock{
			Nonce:                  1,
			AccumulatedFeesInEpoch: expectedAccumulatedFees,
			DevFeesInEpoch:         expectedDevFees,
		}

		result := fees.GetAccumulatedFeesInEpoch(
			epochStartHeader,
			&pool.HeadersPoolStub{},
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, expectedAccumulatedFees, result)

		result = fees.GetDeveloperFeesInEpoch(
			epochStartHeader,
			&pool.HeadersPoolStub{},
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, expectedDevFees, result)
	})

	t.Run("error on GetMetaHeader should return nil (from epoch start meta)", func(t *testing.T) {
		t.Parallel()

		epochStartProposeHash := []byte("epochStartProposeHash")
		epochStartHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 2,
							HeaderHash:  epochStartProposeHash,
						},
					},
				},
			},
		}
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return nil, errors.New("header not found")
			},
		}

		result := fees.GetAccumulatedFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)

		result = fees.GetDeveloperFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)
	})

	t.Run("v3 meta header with one exec result should return nil (exec result of propose epoch start not found)", func(t *testing.T) {
		t.Parallel()

		expectedAccumulatedFees := big.NewInt(4000)
		expectedDevFees := big.NewInt(5000)
		epochStartProposeHash := []byte("epochStartProposeHash")
		epochStartHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 2,
							HeaderHash:  epochStartProposeHash,
						},
					},
				},
			},
		}
		epochStartProposePrevHash := []byte("epochStartProposePrevHash")
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, epochStartProposeHash) {
					return &block.MetaBlockV3{
						EpochChangeProposed: false, // on purpose false
						PrevHash:            epochStartProposePrevHash,
						ExecutionResults: []*block.MetaExecutionResult{
							{
								ExecutionResult: &block.BaseMetaExecutionResult{
									BaseExecutionResult: &block.BaseExecutionResult{
										HeaderNonce: 1,
										HeaderHash:  epochStartProposePrevHash,
									},
									AccumulatedFeesInEpoch: expectedAccumulatedFees,
									DevFeesInEpoch:         expectedDevFees,
								},
							},
						},
					}, nil
				}

				return nil, errors.New("header not found")
			},
		}

		result := fees.GetAccumulatedFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)

		result = fees.GetDeveloperFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, big.NewInt(0), result)
	})

	t.Run("v3 meta header with one exec result should return fees from prev of epoch change proposed", func(t *testing.T) {
		t.Parallel()

		expectedAccumulatedFees := big.NewInt(4000)
		expectedDevFees := big.NewInt(5000)
		epochStartProposeHash := []byte("epochStartProposeHash")
		epochStartHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 2,
							HeaderHash:  epochStartProposeHash,
						},
					},
				},
			},
		}
		epochStartProposePrevHash := []byte("epochStartProposePrevHash")
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, epochStartProposeHash) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
						PrevHash:            epochStartProposePrevHash,
						ExecutionResults: []*block.MetaExecutionResult{
							{
								ExecutionResult: &block.BaseMetaExecutionResult{
									BaseExecutionResult: &block.BaseExecutionResult{
										HeaderNonce: 1,
										HeaderHash:  epochStartProposePrevHash,
									},
									AccumulatedFeesInEpoch: expectedAccumulatedFees,
									DevFeesInEpoch:         expectedDevFees,
								},
							},
						},
					}, nil
				}

				return nil, errors.New("header not found")
			},
		}

		result := fees.GetAccumulatedFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, expectedAccumulatedFees, result)

		result = fees.GetDeveloperFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, expectedDevFees, result)
	})

	t.Run("v3 meta header with multiple exec results should return fees from prev of epoch change proposed", func(t *testing.T) {
		t.Parallel()

		expectedAccumulatedFees := big.NewInt(4000)
		expectedDevFees := big.NewInt(5000)
		epochStartProposeHash := []byte("epochStartProposeHash")
		afterProposedHeaderHash := []byte("afterProposedHeaderHash")
		epochStartHeader := &block.MetaBlockV3{
			Nonce: 4,
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 2,
							HeaderHash:  epochStartProposeHash,
						},
					},
				},
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 3,
							HeaderHash:  afterProposedHeaderHash,
						},
					},
				},
			},
		}
		epochStartProposePrevHash := []byte("epochStartProposePrevHash")
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, epochStartProposeHash) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
						PrevHash:            epochStartProposePrevHash,
						ExecutionResults: []*block.MetaExecutionResult{
							{
								ExecutionResult: &block.BaseMetaExecutionResult{
									BaseExecutionResult: &block.BaseExecutionResult{
										HeaderNonce: 1,
										HeaderHash:  epochStartProposePrevHash,
									},
									AccumulatedFeesInEpoch: expectedAccumulatedFees,
									DevFeesInEpoch:         expectedDevFees,
								},
							},
						},
					}, nil
				}
				if bytes.Equal(hash, afterProposedHeaderHash) {
					return &block.MetaBlockV3{
						PrevHash: epochStartProposeHash,
					}, nil
				}

				return nil, errors.New("header not found")
			},
		}

		result := fees.GetAccumulatedFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, expectedAccumulatedFees, result)

		result = fees.GetDeveloperFeesInEpoch(
			epochStartHeader,
			headersPool,
			&marshallerMock.MarshalizerStub{},
			&storageStubs.ChainStorerStub{},
		)
		require.Equal(t, expectedDevFees, result)
	})
}
