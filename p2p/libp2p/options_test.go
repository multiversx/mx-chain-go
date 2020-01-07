package libp2p

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

//------- WithPeerBlackList

func TestWithPeerBlackList_NilConfigShouldErr(t *testing.T) {
	t.Parallel()

	opt := WithPeerBlackList(&mock.NilBlacklistHandler{})
	err := opt(nil)

	assert.Equal(t, p2p.ErrNilConfigVariable, err)
}

func TestWithPeerBlackList_NilBlackListHadlerShouldErr(t *testing.T) {
	t.Parallel()

	opt := WithPeerBlackList(nil)
	cfg := &p2p.Config{}
	err := opt(cfg)

	assert.Equal(t, p2p.ErrNilPeerBlacklistHandler, err)
}

func TestWithPeerBlackList_ShouldWork(t *testing.T) {
	t.Parallel()

	nblh := &mock.NilBlacklistHandler{}
	opt := WithPeerBlackList(nblh)
	cfg := &p2p.Config{}
	err := opt(cfg)

	assert.Nil(t, err)
	assert.True(t, cfg.BlacklistHandler == nblh)
}
