package peer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainValidatorStatisticsProcessor_ShouldErrNilValidatorStatistics(t *testing.T) {
	t.Parallel()

	scvs, err := peer.NewSovereignChainValidatorStatisticsProcessor(nil)
	assert.Nil(t, scvs)
	assert.Equal(t, process.ErrNilValidatorStatistics, err)
}

func TestNewSovereignChainValidatorStatisticsProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	vs, _ := peer.NewValidatorStatisticsProcessor(args)

	scvs, err := peer.NewSovereignChainValidatorStatisticsProcessor(vs)
	assert.NotNil(t, scvs)
	assert.Nil(t, err)
}

func TestUpdateShardDataPeerState_ShouldReturnNil(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	vs, _ := peer.NewValidatorStatisticsProcessor(args)
	scvs, _ := peer.NewSovereignChainValidatorStatisticsProcessor(vs)

	err := scvs.UpdateShardDataPeerState(nil, nil)
	assert.Nil(t, err)
}
