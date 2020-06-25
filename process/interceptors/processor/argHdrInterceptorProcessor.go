package processor

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgHdrInterceptorProcessor is the argument for the interceptor processor used for headers (shard, meta and so on)
type ArgHdrInterceptorProcessor struct {
	Headers        dataRetriever.HeadersPool
	HdrValidator   process.HeaderValidator
	BlockBlackList process.TimeCacher
}
