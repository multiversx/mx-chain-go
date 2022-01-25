package blockAPI

import "github.com/ElrondNetwork/elrond-go-core/core"

// CreateAPIBlockProcessor will create a new instance of APIBlockHandler
func CreateAPIBlockProcessor(args *APIBlockProcessorArg) (APIBlockHandler, error) {
	if args.SelfShardID != core.MetachainShardId {
		return NewShardApiBlockProcessor(args), nil
	}

	return NewMetaApiBlockProcessor(args), nil
}
