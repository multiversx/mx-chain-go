package p2p_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

func TestPeerDiscoveryType_StringOffUsingMixedCaseShouldWork(t *testing.T) {
	t.Parallel()

	pdt, err := p2p.LoadPeerDiscoveryTypeFromString("oFf")

	assert.Equal(t, p2p.PeerDiscoveryOff, pdt)
	assert.Nil(t, err)
}

func TestPeerDiscoveryType_StringMdnsUsingMixedCaseShouldWork(t *testing.T) {
	t.Parallel()

	pdt, err := p2p.LoadPeerDiscoveryTypeFromString("mDnS")

	assert.Equal(t, p2p.PeerDiscoveryMdns, pdt)
	assert.Nil(t, err)
}

func TestPeerDiscoveryType_StringKadDhtUsingMixedCaseShouldWork(t *testing.T) {
	t.Parallel()

	pdt, err := p2p.LoadPeerDiscoveryTypeFromString("kAd-DhT")

	assert.Equal(t, p2p.PeerDiscoveryKadDht, pdt)
	assert.Nil(t, err)
}

func TestPeerDiscoveryType_StringNotImplementedShouldErr(t *testing.T) {
	t.Parallel()

	_, err := p2p.LoadPeerDiscoveryTypeFromString("not-implemented-type")

	assert.Equal(t, p2p.ErrPeerDiscoveryNotImplemented, err)
}
