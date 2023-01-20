package process

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestIdentifiers_ShardCacherIdentifier(t *testing.T) {
	require.Equal(t, "0", ShardCacherIdentifier(0, 0))
	require.Equal(t, "4", ShardCacherIdentifier(4, 4))
	require.Equal(t, "0_1", ShardCacherIdentifier(0, 1))
	require.Equal(t, "1_0", ShardCacherIdentifier(1, 0))
}

func TestIdentifiers_ParseShardCacherIdentifier(t *testing.T) {
	assertParseShardCacherIdentifier(t, "", 0, 0, ErrInvalidShardCacherIdentifier)
	assertParseShardCacherIdentifier(t, "_", 0, 0, ErrInvalidShardCacherIdentifier)
	assertParseShardCacherIdentifier(t, "?", 0, 0, ErrInvalidShardCacherIdentifier)
	assertParseShardCacherIdentifier(t, "0", 0, 0, nil)
	assertParseShardCacherIdentifier(t, "1", 1, 1, nil)
	assertParseShardCacherIdentifier(t, "0_1", 0, 1, nil)
	assertParseShardCacherIdentifier(t, "1_0", 1, 0, nil)
	assertParseShardCacherIdentifier(t, "2_2", 2, 2, nil)
	assertParseShardCacherIdentifier(t, "4_3", 4, 3, nil)
	assertParseShardCacherIdentifier(t, "0_4294967295", 0, core.MetachainShardId, nil)
}

func assertParseShardCacherIdentifier(t *testing.T, cacheID string, source uint32, destination uint32, err error) {
	actoualSource, actualDestination, actualErr := ParseShardCacherIdentifier(cacheID)
	require.Equal(t, source, actoualSource)
	require.Equal(t, destination, actualDestination)
	require.Equal(t, err, actualErr)
}

func TestIdentifiers_IsShardCacherIdentifierForSourceMe(t *testing.T) {
	require.True(t, IsShardCacherIdentifierForSourceMe("0", 0))
	require.True(t, IsShardCacherIdentifierForSourceMe("1", 1))
	require.False(t, IsShardCacherIdentifierForSourceMe("2", 1))
	require.False(t, IsShardCacherIdentifierForSourceMe("2_2", 2)) // Bad format
	require.False(t, IsShardCacherIdentifierForSourceMe("", 0))
	require.False(t, IsShardCacherIdentifierForSourceMe("2_3", 2))
}
