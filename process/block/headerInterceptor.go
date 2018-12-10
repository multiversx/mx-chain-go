package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

type headerInterceptor struct {
	intercept  *process.Interceptor
	headerPool *dataPool.DataPool
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	messenger p2p.Messenger,
	headerPool *dataPool.DataPool,
) (*headerInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if headerPool == nil {
		return nil, process.ErrNilHeaderDataPool
	}

	intercept, err := process.NewInterceptor("hdr", messenger, NewInterceptedHeader())
	if err != nil {
		return nil, err
	}

	hdrIntercept := &headerInterceptor{
		intercept:  intercept,
		headerPool: headerPool,
	}

	intercept.CheckReceivedObject = hdrIntercept.processHdr

	return hdrIntercept, nil
}

func (hi *headerInterceptor) processHdr(hdr p2p.Newer, rawData []byte, hasher hashing.Hasher) bool {
	if hdr == nil {
		log.Debug("nil hdr to process")
		return false
	}

	if hasher == nil {
		return false
	}

	hdrIntercepted, ok := hdr.(process.HeaderInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: headerInterceptor is not using InterceptedHeader " +
			"as template object and will always return false")
		return false
	}

	hash := hasher.Compute(string(rawData))
	hdrIntercepted.SetHash(hash)

	if !hdrIntercepted.Check() || !hdrIntercepted.VerifySig() {
		return false
	}

	hi.headerPool.AddData(hash, hdrIntercepted, hdrIntercepted.Shard())
	return true
}
