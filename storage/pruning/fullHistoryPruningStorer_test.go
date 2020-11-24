package pruning_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/stretchr/testify/assert"
)

func TestNewFullHistoryPruningStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 10,
	}
	ps, err := pruning.NewFullHistoryPruningStorer(fhArgs)

	assert.NotNil(t, ps)
	assert.Nil(t, err)
	assert.False(t, ps.IsInterfaceNil())
}
