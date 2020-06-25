package peerHonesty

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type peerScore struct {
	pk            string
	scoresByTopic map[string]float64
}

func newPeerScore(pk string) *peerScore {
	return &peerScore{
		pk:            pk,
		scoresByTopic: make(map[string]float64),
	}
}

func (ps *peerScore) size() int {
	return len(ps.pk) + len(ps.scoresByTopic)*(float64Size+defaultTopicSize)
}

// String will return the peerScore contents in a string - not concurrent safe
func (ps *peerScore) String() string {
	scores := make([]string, 0, len(ps.scoresByTopic))
	for topic, score := range ps.scoresByTopic {
		scores = append(scores, fmt.Sprintf("%s: %.2f", topic, score))
	}

	return fmt.Sprintf("%s scoring: %s", hex.EncodeToString([]byte(ps.pk)), strings.Join(scores, ", "))
}
