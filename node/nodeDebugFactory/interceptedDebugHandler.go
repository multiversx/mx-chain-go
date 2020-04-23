package nodeDebugFactory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/debug/factory"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorResolverDebugger is the contant string for the debugger
const InterceptorResolverDebugger = "interceptor resolver debugger"

// CreateInterceptedDebugHandler creates and applies an interceptor-resolver debug handler
func CreateInterceptedDebugHandler(
	node NodeWrapper,
	interceptors process.InterceptorsContainer,
	resolvers dataRetriever.ResolversFinder,
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

	debugHandler, err := factory.NewInterceptorResolverDebuggerFactory(config)
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
		err = resolver.SetResolverDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on resolvers", errFound)
	}

	return node.AddQueryHandler(InterceptorResolverDebugger, debugHandler)
}
