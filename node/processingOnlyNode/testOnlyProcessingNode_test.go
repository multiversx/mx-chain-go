package processingOnlyNode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createMockArgsTestOnlyProcessingNode() ArgsTestOnlyProcessingNode {
	return ArgsTestOnlyProcessingNode{
		NumShards: 0,
		ShardID:   3,
	}
}

func TestNewTestOnlyProcessingNode(t *testing.T) {
	t.Parallel()

	t.Run("invalid shard configuration should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode()
		args.ShardID = args.NumShards
		node, err := NewTestOnlyProcessingNode(args)
		assert.NotNil(t, err)
		assert.Nil(t, node)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode()
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})
}
