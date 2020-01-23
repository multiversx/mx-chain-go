package factory

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("process/interceptors/factory")

// nilFinalityAttester is a nil implementation that will cause all received headers
//not to be actually checked
type nilFinalityAttester struct {
}

// GetFinalHeader returns a header handler implementation having the same shardID as the one provided
// a 0 value for nonce. The returned hash is an empty slice and the error is nil
func (nfa *nilFinalityAttester) GetFinalHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	log.Warn("nilFinalityAttester.GetFinalHeader", "usage", "should not have been used")
	hdr := &block.Header{
		Nonce:   0,
		ShardId: shardID,
	}

	return hdr, make([]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nfa *nilFinalityAttester) IsInterfaceNil() bool {
	return nfa == nil
}
