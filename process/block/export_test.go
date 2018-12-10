package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (hi *headerInterceptor) ProcessHdr(hdr p2p.Newer, rawData []byte, hasher hashing.Hasher) bool {
	return hi.processHdr(hdr, rawData, hasher)
}
