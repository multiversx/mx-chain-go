package blockAPI

import "github.com/ElrondNetwork/elrond-go-core/core"

// CreateAPIBlockProcessor will create a new instance of APIBlockHandler
func CreateAPIBlockProcessor(arg *ArgAPIBlockProcessor) (APIBlockHandler, error) {
	err := checkNilArg(arg)
	if err != nil {
		return nil, err
	}

	if arg.SelfShardID != core.MetachainShardId {
		return NewShardApiBlockProcessor(arg), nil
	}

	return NewMetaApiBlockProcessor(arg), nil
}
