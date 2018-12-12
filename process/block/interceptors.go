package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// HeaderInterceptor represents an interceptor used for block headers
type HeaderInterceptor struct {
	intercept     *interceptor.Interceptor
	headers       data.ShardedDataCacherNotifier
	headersNonces data.Uint64Cacher
	hasher        hashing.Hasher
}

// GenericBlockBodyInterceptor represents an interceptor used for all types of block bodies
type GenericBlockBodyInterceptor struct {
	intercept *interceptor.Interceptor
	cache     storage.Cacher
	hasher    hashing.Hasher
}

//------- HeaderInterceptor

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	messenger p2p.Messenger,
	headers data.ShardedDataCacherNotifier,
	headersNonces data.Uint64Cacher,
	hasher hashing.Hasher,
) (*HeaderInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
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
		intercept:     intercept,
		headers:       headers,
		headersNonces: headersNonces,
		hasher:        hasher,
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

	hi.headers.AddData(hash, hdrIntercepted, hdrIntercepted.Shard())
	if hi.checkHeaderForCurrentShard(hdrIntercepted) {
		_, _ = hi.headersNonces.HasOrAdd(hdrIntercepted.GetHeader().Nonce, hash)
	}
	return true
}

func (hi *HeaderInterceptor) checkHeaderForCurrentShard(header process.HeaderInterceptorAdapter) bool {
	//TODO add real logic here
	return true
}

//------- GenericBlockBodyInterceptor

// NewGenericBlockBodyInterceptor hooks a new interceptor for block bodies
// Fetched data blocks will be placed inside the cache
func NewGenericBlockBodyInterceptor(
	name string,
	messenger p2p.Messenger,
	cache storage.Cacher,
	hasher hashing.Hasher,
	templateObj process.BlockBodyInterceptorAdapter,
) (*GenericBlockBodyInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if cache == nil {
		return nil, process.ErrNilCacher
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if templateObj == nil {
		return nil, process.ErrNilTemplateObj
	}

	intercept, err := interceptor.NewInterceptor(name, messenger, templateObj)
	if err != nil {
		return nil, err
	}

	bbIntercept := &GenericBlockBodyInterceptor{
		intercept: intercept,
		cache:     cache,
		hasher:    hasher,
	}

	intercept.CheckReceivedObject = bbIntercept.processBodyBlock

	return bbIntercept, nil
}

func (gbbi *GenericBlockBodyInterceptor) processBodyBlock(bodyBlock p2p.Newer, rawData []byte) bool {
	if bodyBlock == nil {
		log.Debug("nil body block to process")
		return false
	}

	txBlockBodyIntercepted, ok := bodyBlock.(process.BlockBodyInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: BlockBodyInterceptor is not using BlockBodyInterceptorAdapter " +
			"as template object and will always return false")
		return false
	}

	hash := gbbi.hasher.Compute(string(rawData))
	txBlockBodyIntercepted.SetHash(hash)

	if !txBlockBodyIntercepted.Check() {
		return false
	}

	_ = gbbi.cache.Put(hash, txBlockBodyIntercepted)
	return true
}
