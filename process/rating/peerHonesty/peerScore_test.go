package peerHonesty

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerScore_String(t *testing.T) {
	t.Parallel()

	ps := newPeerScore("pk")
	topic1 := "topic1"
	topic2 := "topic2"

	ps.scoresByTopic[topic1] = 1.2
	ps.scoresByTopic[topic2] = 1.3

	str := ps.String()

	assert.True(t, strings.Contains(str, topic1))
	assert.True(t, strings.Contains(str, topic2))
}
