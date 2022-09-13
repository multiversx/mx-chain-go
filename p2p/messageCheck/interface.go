package messagecheck

import "github.com/ElrondNetwork/elrond-go-core/core"

type p2pSigner interface {
	Verify(payload []byte, pid core.PeerID, signature []byte) error
}
