package slash

import "github.com/ElrondNetwork/elrond-go-core/data"

// HeaderList defines a list of data.HeaderHandler
type HeaderList []data.HeaderHandler

// HeaderInfoList defines a list of data.HeaderInfoHandler
type HeaderInfoList []data.HeaderInfoHandler

// HeaderInfo defines a HeaderHandler and its corresponding hash
type HeaderInfo struct {
	Header data.HeaderHandler
	Hash   []byte
}

// GetHeaderHandler returns data.HeaderHandler
func (hi *HeaderInfo) GetHeaderHandler() data.HeaderHandler {
	return hi.Header
}

// GetHash returns header's hash
func (hi *HeaderInfo) GetHash() []byte {
	return hi.Hash
}
