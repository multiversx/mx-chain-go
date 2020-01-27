package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func initRoundConsensus() *spos.RoundConsensus {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		len(eligibleList),
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.ResetRoundState()

	return spos.NewRoundConsensusWrapper(rcns)
}

func TestRoundConsensus_NewRoundConsensusShouldWork(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.NotNil(t, rcns)
	assert.Equal(t, 3, len(rcns.ConsensusGroup()))
	assert.Equal(t, "3", rcns.ConsensusGroup()[2])
	assert.Equal(t, "2", rcns.SelfPubKey())
}

func TestRoundConsensus_ConsensusGroupIndexFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key3")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key3")

	assert.Equal(t, 2, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_ConsensusGroupIndexNotFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key4")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key4")

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupInConsesus(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key2")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.SelfConsensusGroupIndex()

	assert.Equal(t, 1, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupNotFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key4")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.SelfConsensusGroupIndex()

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestRoundConsensus_SetEligibleListShouldChangeTheEligibleList(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetEligibleList([]string{"4", "5", "6"})

	assert.Equal(t, "4", rcns.EligibleList()[0])
	assert.Equal(t, "5", rcns.EligibleList()[1])
	assert.Equal(t, "6", rcns.EligibleList()[2])
}

func TestRoundConsensus_SetConsensusGroupShouldChangeTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetConsensusGroup([]string{"4", "5", "6"})

	assert.Equal(t, "4", rcns.ConsensusGroup()[0])
	assert.Equal(t, "5", rcns.ConsensusGroup()[1])
	assert.Equal(t, "6", rcns.ConsensusGroup()[2])
}

func TestRoundConsensus_SetConsensusGroupSizeShouldChangeTheConsensusGroupSize(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, len(rcns.ConsensusGroup()), rcns.ConsensusGroupSize())
	rcns.SetConsensusGroupSize(99999)
	assert.Equal(t, 99999, rcns.ConsensusGroupSize())
}

func TestRoundConsensus_SetSelfPubKeyShouldChangeTheSelfPubKey(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetSelfPubKey("X")
	assert.Equal(t, "X", rcns.SelfPubKey())
}

func TestRoundConsensus_IsNodeInConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, false, rcns.IsNodeInConsensusGroup("4"))
	assert.Equal(t, true, rcns.IsNodeInConsensusGroup(rcns.SelfPubKey()))
}

func TestRoundConsensus_IsNodeInEligibleList(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, false, rcns.IsNodeInEligibleList("4"))
	assert.Equal(t, true, rcns.IsNodeInEligibleList(rcns.SelfPubKey()))
}
