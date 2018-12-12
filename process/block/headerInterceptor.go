package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
)

var log = logger.NewDefaultLogger()

// HeaderInterceptor represents an interceptor used for block headers
type HeaderInterceptor struct {
	intercept  *interceptor.Interceptor
	headerPool *shardedData.ShardedData
	hasher     hashing.Hasher
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	messenger p2p.Messenger,
	headerPool *shardedData.ShardedData,
	hasher hashing.Hasher,
) (*HeaderInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if headerPool == nil {
		return nil, process.ErrNilHeaderDataPool
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	intercept, err := interceptor.NewInterceptor(
		process.HeaderInterceptor,
		messenger,
		NewInterceptedHeader())
	if err != nil {
		return nil, err
	}

	hdrIntercept := &HeaderInterceptor{
		intercept:  intercept,
		headerPool: headerPool,
		hasher:     hasher,
	}

	intercept.CheckReceivedObject = hdrIntercept.processHdr

	return hdrIntercept, nil
}

func (hi *HeaderInterceptor) processHdr(hdr p2p.Newer, rawData []byte) bool {
	if hdr == nil {
		log.Debug("nil hdr to process")
		return false
	}

	hdrIntercepted, ok := hdr.(process.HeaderInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: headerInterceptor is not using InterceptedHeader " +
			"as template object and will always return false")
		return false
	}

	hash := hi.hasher.Compute(string(rawData))
	hdrIntercepted.SetHash(hash)

	if !hdrIntercepted.Check() || !hdrIntercepted.VerifySig() {
		return false
	}

	hi.headerPool.AddData(hash, hdrIntercepted, hdrIntercepted.Shard())
	return true
}
