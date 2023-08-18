package groups

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestExtractAccountQueryOptions(t *testing.T) {
	t.Parallel()

	t.Run("good options", func(t *testing.T) {
		t.Parallel()

		options, err := extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onFinalBlock=true"))
		require.Nil(t, err)
		require.True(t, options.OnFinalBlock)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onStartOfEpoch=7"))
		require.Nil(t, err)
		require.Equal(t, core.OptionalUint32{Value: 7, HasValue: true}, options.OnStartOfEpoch)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockNonce=42"))
		require.Nil(t, err)
		require.Equal(t, core.OptionalUint64{Value: 42, HasValue: true}, options.BlockNonce)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockHash=aaaa"))
		require.Nil(t, err)
		require.Equal(t, []byte{0xaa, 0xaa}, options.BlockHash)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockHash=aaaa"))
		require.Nil(t, err)
		require.Equal(t, []byte{0xaa, 0xaa}, options.BlockHash)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockRootHash=bbbb&hintEpoch=7"))
		require.Nil(t, err)
		require.Equal(t, []byte{0xbb, 0xbb}, options.BlockRootHash)
		require.Equal(t, uint32(7), options.HintEpoch.Value)
	})

	t.Run("bad options", func(t *testing.T) {
		t.Parallel()

		options, err := extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockNonce=42&blockHash=aaaa"))
		require.ErrorContains(t, err, "only one block coordinate")
		require.Equal(t, api.AccountQueryOptions{}, options)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockHash=aaaa&blockRootHash=bbbb"))
		require.ErrorContains(t, err, "only one block coordinate")
		require.Equal(t, api.AccountQueryOptions{}, options)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onFinalBlock=true&blockHash=aaaa"))
		require.ErrorContains(t, err, "onFinalBlock is not compatible")
		require.Equal(t, api.AccountQueryOptions{}, options)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onStartOfEpoch=7&blockRootHash=bbbb"))
		require.ErrorContains(t, err, "onStartOfEpoch is not compatible")
		require.Equal(t, api.AccountQueryOptions{}, options)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onFinalBlock=true&hintEpoch=7"))
		require.ErrorContains(t, err, "hintEpoch is optional, but only compatible with blockRootHash")
		require.Equal(t, api.AccountQueryOptions{}, options)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockHash=aaaa&hintEpoch=7"))
		require.ErrorContains(t, err, "hintEpoch is optional, but only compatible with blockRootHash")
		require.Equal(t, api.AccountQueryOptions{}, options)

		options, err = extractAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockNonce=aaaa"))
		require.ErrorContains(t, err, errors.ErrBadUrlParams.Error())
		require.Equal(t, api.AccountQueryOptions{}, options)
	})
}

func TestParseAccountQueryOptions(t *testing.T) {
	t.Parallel()

	options, err := parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onFinalBlock=test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
	require.Equal(t, api.AccountQueryOptions{}, options)

	options, err = parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("onStartOfEpoch=test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
	require.Equal(t, api.AccountQueryOptions{}, options)

	options, err = parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockNonce=test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
	require.Equal(t, api.AccountQueryOptions{}, options)

	options, err = parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockHash=test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid byte")
	require.Equal(t, api.AccountQueryOptions{}, options)

	options, err = parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("blockRootHash=test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid byte")
	require.Equal(t, api.AccountQueryOptions{}, options)

	options, err = parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery("hintEpoch=test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
	require.Equal(t, api.AccountQueryOptions{}, options)

	options, err = parseAccountQueryOptions(testscommon.CreateGinContextWithRawQuery(""))
	require.NoError(t, err)
	require.Equal(t, api.AccountQueryOptions{}, options)
}
