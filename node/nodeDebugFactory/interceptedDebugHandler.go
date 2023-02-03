package nodeDebugFactory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/debug/factory"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptorDebugger is the constant string for the debugger
const InterceptorDebugger = "interceptor debugger"

// CreateInterceptedDebugHandler creates and applies an interceptor-resolver debug handler
func CreateInterceptedDebugHandler(
	node NodeWrapper,
	interceptors process.InterceptorsContainer,
	resolvers dataRetriever.ResolversContainer,
	requesters dataRetriever.RequestersContainer,
	config config.InterceptorResolverDebugConfig,
) error {
	if check.IfNil(node) {
		return ErrNilNodeWrapper
	}
	if check.IfNil(interceptors) {
		return ErrNilInterceptorContainer
	}
	if check.IfNil(resolvers) {
		return ErrNilResolverContainer
	}
	if check.IfNil(requesters) {
		return ErrNilRequestersContainer
	}

	debugHandler, err := factory.NewInterceptorDebuggerFactory(config)
	if err != nil {
		return err
	}

	var errFound error
	interceptors.Iterate(func(key string, interceptor process.Interceptor) bool {
		err = interceptor.SetInterceptedDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on interceptors", errFound)
	}

	resolvers.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		err = resolver.SetDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on resolvers", errFound)
	}

	requesters.Iterate(func(key string, requester dataRetriever.Requester) bool {
		err = requester.SetDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on resolvers", errFound)
	}

	return node.AddQueryHandler(InterceptorDebugger, debugHandler)
}
