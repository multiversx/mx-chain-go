package processor

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgHdrInterceptorProcessor is the argument for the interceptor processor used for headers (shard, meta and so on)
type ArgHdrInterceptorProcessor struct {
	Headers             dataRetriever.HeadersPool
	Proofs              dataRetriever.ProofsPool
	BlockBlackList      process.TimeCacher
	EnableEpochsHandler common.EnableEpochsHandler
}
