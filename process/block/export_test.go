package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

func (hi *HeaderInterceptor) ProcessHdr(hdr p2p.Newer, rawData []byte) bool {
	return hi.processHdr(hdr, rawData)
}

func (gbbi *GenericBlockBodyInterceptor) ProcessBodyBlock(bodyBlock p2p.Newer, rawData []byte) bool {
	return gbbi.processBodyBlock(bodyBlock, rawData)
}

func (hdrRes *HeaderResolver) ResolveHdrRequest(rd process.RequestData) ([]byte, error) {
	return hdrRes.resolveHdrRequest(rd)
}

func (bbRes *BlockBodyResolver) ResolveBlockBodyRequest(rd process.RequestData) ([]byte, error) {
	return bbRes.resolveBlockBodyRequest(rd)
}
