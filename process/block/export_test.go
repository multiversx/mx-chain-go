package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (hi *HeaderInterceptor) ProcessHdr(hdr p2p.Newer, rawData []byte) bool {
	return hi.processHdr(hdr, rawData)
}

func (gbbi *GenericBlockBodyInterceptor) ProcessBodyBlock(bodyBlock p2p.Newer, rawData []byte) bool {
	return gbbi.processBodyBlock(bodyBlock, rawData)
}
