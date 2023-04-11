package chainparametersnotifier

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("common/chainparameters")

type chainParametersNotifier struct {
	mutData                sync.RWMutex
	wasInitialized         bool
	currentChainParameters config.ChainParametersByEpochConfig
	mutHandler             sync.RWMutex
	handlers               []common.ChainParametersSubscriptionHandler
}

// NewChainParametersNotifier creates a new instance of a chainParametersNotifier component
func NewChainParametersNotifier() *chainParametersNotifier {
	return &chainParametersNotifier{
		wasInitialized: false,
		handlers:       make([]common.ChainParametersSubscriptionHandler, 0),
	}
}

// UpdateCurrentChainParameters should be called whenever new chain parameters become active on the network
func (cpn *chainParametersNotifier) UpdateCurrentChainParameters(params config.ChainParametersByEpochConfig) {
	cpn.mutData.Lock()
	shouldSkipParams := cpn.wasInitialized && cpn.currentChainParameters.EnableEpoch == params.EnableEpoch
	if shouldSkipParams {
		cpn.mutData.Unlock()

		return
	}
	cpn.wasInitialized = true
	cpn.currentChainParameters = params
	cpn.mutData.Unlock()

	cpn.mutHandler.RLock()
	handlersCopy := make([]common.ChainParametersSubscriptionHandler, len(cpn.handlers))
	copy(handlersCopy, cpn.handlers)
	cpn.mutHandler.RUnlock()

	log.Debug("chainParametersNotifier.UpdateCurrentChainParameters",
		"enable epoch", params.EnableEpoch,
		"shard consensus group size", params.ShardConsensusGroupSize,
		"shard min number of nodes", params.ShardMinNumNodes,
		"meta consensus group size", params.MetachainConsensusGroupSize,
		"meta min number of nodes", params.MetachainMinNumNodes,
		"round duration", params.RoundDuration,
		"hysteresis", params.Hysteresis,
		"adaptivity", params.Adaptivity,
	)

	for _, handler := range handlersCopy {
		handler.ChainParametersChanged(params)
	}
}

// RegisterNotifyHandler will register the provided handler to be called whenever chain parameters have changed
func (cpn *chainParametersNotifier) RegisterNotifyHandler(handler common.ChainParametersSubscriptionHandler) {
	if check.IfNil(handler) {
		return
	}

	cpn.mutHandler.Lock()
	cpn.handlers = append(cpn.handlers, handler)
	cpn.mutHandler.Unlock()

	cpn.mutData.RLock()
	handler.ChainParametersChanged(cpn.currentChainParameters)
	cpn.mutData.RUnlock()
}

// CurrentChainParameters returns the current chain parameters
func (cpn *chainParametersNotifier) CurrentChainParameters() config.ChainParametersByEpochConfig {
	cpn.mutData.RLock()
	defer cpn.mutData.RUnlock()

	return cpn.currentChainParameters
}

// UnRegisterAll removes all registered handlers queue
func (cpn *chainParametersNotifier) UnRegisterAll() {
	cpn.mutHandler.Lock()
	cpn.handlers = make([]common.ChainParametersSubscriptionHandler, 0)
	cpn.mutHandler.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (cpn *chainParametersNotifier) IsInterfaceNil() bool {
	return cpn == nil
}
