package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
)

func TestEpochStartBootstrap_MultipleEpochOffset(t *testing.T) {
	t.Parallel()

	for _, offset := range []uint32{2, 5, 25} {
		t.Run(fmt.Sprintf("offset %d", offset), func(t *testing.T) {
			t.Parallel()
			const latestEpoch = uint32(30)
			headers := make(map[string]data.HeaderHandler)
			for epoch := uint32(1); epoch <= latestEpoch; epoch++ {
				headers[fmt.Sprint(epoch)] = createEpochStartMetaForOffsetTest(epoch, []byte(fmt.Sprint(epoch-1)))
			}
			provider := createEpochStartBootstrapForOffsetTest(t, headers["30"].(data.MetaHeaderHandler))
			provider.flagsConfig.StartInEpochOffset = offset
			var requestEpoch uint32
			var requestedEpochs []uint32
			var requestedHash string
			provider.requestHandler = &testscommon.RequestHandlerStub{
				SetEpochCalled: func(epoch uint32) { requestEpoch = epoch },
			}
			provider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
				SyncMissingHeadersByHashCalled: func(shards []uint32, hashes [][]byte, _ context.Context) error {
					require.Equal(t, []uint32{core.MetachainShardId}, shards)
					require.Equal(t, [][]byte{[]byte(fmt.Sprint(requestEpoch))}, hashes)
					requestedEpochs = append(requestedEpochs, requestEpoch)
					requestedHash = string(hashes[0])
					return nil
				},
				GetHeadersCalled: func() (map[string]data.HeaderHandler, error) {
					return map[string]data.HeaderHandler{requestedHash: headers[requestedHash]}, nil
				},
			}

			_, shouldReturn, err := provider.applyStartInEpochOffset()
			require.NoError(t, err)
			require.False(t, shouldReturn)
			require.Len(t, requestedEpochs, int(offset))
			for i, epoch := range requestedEpochs {
				require.Equal(t, latestEpoch-1-uint32(i), epoch)
			}
			require.Equal(t, latestEpoch-offset, requestEpoch)
			require.Same(t, headers[fmt.Sprint(latestEpoch-offset)], provider.epochStartMeta)
		})
	}
}

func TestEpochStartBootstrap_OffsetBoundaries(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                  string
		latest, start, offset uint32
		invalid               bool
	}{
		{"genesis", 25, 0, 25, false},
		{"hardfork start", 30, 7, 23, false},
		{"before genesis", 25, 0, 26, true},
		{"before hardfork", 30, 7, 24, true},
		{"latest below start", 6, 7, 1, true},
		{"maximum offset underflow", 30, 0, math.MaxUint32, true},
		{"maximum epoch to genesis", math.MaxUint32, 0, math.MaxUint32, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			latest := createEpochStartMetaForOffsetTest(tc.latest, []byte("unused"))
			provider := createEpochStartBootstrapForOffsetTest(t, latest)
			provider.startEpoch = tc.start
			provider.flagsConfig.StartInEpochOffset = tc.offset
			provider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
				SyncMissingHeadersByHashCalled: func(_ []uint32, _ [][]byte, _ context.Context) error {
					t.Fatal("boundary cases must not request historical headers")
					return nil
				},
			}
			params, shouldReturn, err := provider.applyStartInEpochOffset()
			if tc.invalid {
				require.ErrorIs(t, err, epochStart.ErrInvalidStartInEpochOffset)
				require.False(t, shouldReturn)
			} else {
				require.NoError(t, err)
				require.True(t, shouldReturn)
				require.Equal(t, tc.start, params.Epoch)
			}
			require.Same(t, latest, provider.epochStartMeta)
		})
	}
}

func TestEpochStartBootstrap_OffsetIntermediateFailure(t *testing.T) {
	t.Parallel()
	missingHistory := errors.New("historical header or proof unavailable")
	for _, tc := range []struct {
		name             string
		header           data.HeaderHandler
		syncErr, wantErr error
	}{
		{"unavailable history", nil, missingHistory, missingHistory},
		{"broken epoch sequence", createEpochStartMetaForOffsetTest(1, []byte("genesis")), nil, epochStart.ErrEpochStartMetaBlockEpochMismatch},
		{"missing header", nil, nil, epochStart.ErrMissingHeader},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			latest := createEpochStartMetaForOffsetTest(5, []byte("four"))
			provider := createEpochStartBootstrapForOffsetTest(t, latest)
			provider.flagsConfig.StartInEpochOffset = 3
			requests := 0
			provider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
				SyncMissingHeadersByHashCalled: func(_ []uint32, _ [][]byte, _ context.Context) error {
					requests++
					if requests == 2 {
						return tc.syncErr
					}
					return nil
				},
				GetHeadersCalled: func() (map[string]data.HeaderHandler, error) {
					if requests == 1 {
						return map[string]data.HeaderHandler{"four": createEpochStartMetaForOffsetTest(4, []byte("three"))}, nil
					}
					return map[string]data.HeaderHandler{"three": tc.header}, nil
				},
			}
			_, shouldReturn, err := provider.applyStartInEpochOffset()
			require.ErrorIs(t, err, tc.wantErr)
			require.False(t, shouldReturn)
			require.Equal(t, 2, requests)
			require.Same(t, latest, provider.epochStartMeta, "must not select a partially traversed epoch")
		})
	}
}
