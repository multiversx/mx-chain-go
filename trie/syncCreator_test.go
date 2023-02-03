package trie

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrieSync_InvalidVersionShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	syncer, err := CreateTrieSyncer(arg, 0)

	assert.True(t, check.IfNil(syncer))
	assert.True(t, errors.Is(err, ErrInvalidTrieSyncerVersion))
}

func TestNewTrieSync_FirstVariantImplementation(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	syncer, err := CreateTrieSyncer(arg, initialVersion)

	require.False(t, check.IfNil(syncer))
	require.Nil(t, err)
	_, isInstanceOk := syncer.(*trieSyncer)
	assert.True(t, isInstanceOk)
}

func TestNewTrieSync_SecondVariantImplementation(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	syncer, err := CreateTrieSyncer(arg, secondVersion)

	require.False(t, check.IfNil(syncer))
	require.Nil(t, err)
	_, isInstanceOk := syncer.(*doubleListTrieSyncer)
	assert.True(t, isInstanceOk)
}

func TestNewTrieSync_ThirdVariantImplementation(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	syncer, err := CreateTrieSyncer(arg, thirdVersion)

	require.False(t, check.IfNil(syncer))
	require.Nil(t, err)
	_, isInstanceOk := syncer.(*depthFirstTrieSyncer)
	assert.True(t, isInstanceOk)
}

func TestCheckTrieSyncerVersion(t *testing.T) {
	t.Parallel()

	err := CheckTrieSyncerVersion(0)
	assert.True(t, errors.Is(err, ErrInvalidTrieSyncerVersion))

	err = CheckTrieSyncerVersion(initialVersion)
	assert.Nil(t, err)

	err = CheckTrieSyncerVersion(secondVersion)
	assert.Nil(t, err)

	err = CheckTrieSyncerVersion(thirdVersion)
	assert.Nil(t, err)

	err = CheckTrieSyncerVersion(4)
	assert.True(t, errors.Is(err, ErrInvalidTrieSyncerVersion))
}
