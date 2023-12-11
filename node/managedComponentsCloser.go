package node

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/factory"
)

type managedComponentsCloser struct {
	validComponentsList []string
	allComponents       map[string]factory.ComponentHandler
}

func newManagedComponentsCloser() *managedComponentsCloser {
	return &managedComponentsCloser{
		validComponentsList: getClosingOrderForManagedComponents(),
		allComponents:       make(map[string]factory.ComponentHandler),
	}
}

func (closerHandler *managedComponentsCloser) addManagedComponent(component factory.ComponentHandler) error {
	if check.IfNilReflect(component) {
		return fmt.Errorf("programming error: %w, stack: %s", ErrNilComponentHandler, string(debug.Stack()))
	}

	componentName := component.String()
	if !find(componentName, closerHandler.validComponentsList) {
		return fmt.Errorf("programming error: %w, identifier: %s, stack: %s", ErrInvalidComponentHandler, componentName, string(debug.Stack()))
	}

	_, exists := closerHandler.allComponents[componentName]
	if exists {
		return fmt.Errorf("programming error: %w, identifier: %s, stack: %s",
			ErrDuplicatedComponentHandler, componentName, string(debug.Stack()))
	}

	closerHandler.allComponents[componentName] = component

	return nil
}

func getClosingOrderForManagedComponents() []string {
	return []string{
		factory.NetworkComponentsName,     // first to be closed so the node won't be able to send wrong p2p data because of the closing process
		factory.ConsensusComponentsName,   // highest level component order
		factory.HeartbeatV2ComponentsName, // doesn't quite matter, we close it here, in the first part
		factory.ProcessComponentsName,     // process components are very high level components
		factory.StatusComponentsName,      // status components can be safely closed here

		// TODO - just for testing
		factory.DataComponentsName,  // close the data components now because nobody should be still using it at this point
		factory.StateComponentsName, // close the state components after the processing is done and before the data components

		factory.BootstrapComponentsName,  // close the first processing components of the node used in the startup phase of the node
		factory.CryptoComponentsName,     // nobody should still use the crypto components now
		factory.CoreComponentsName,       // core components can be safely closed now
		factory.StatusCoreComponentsName, // status core components were the first to be created, let's close them
	}
}

// close will call Close on the provided components in a predefined order.
// The order is important, always analyze before changing this function
func (closerHandler *managedComponentsCloser) close() error {
	log.Debug("closing all managed components", "full components list, in order", strings.Join(closerHandler.validComponentsList, ", "))

	var closeError error = nil

	for _, componentName := range closerHandler.validComponentsList {
		component, componentExists := closerHandler.allComponents[componentName]
		if !componentExists {
			log.Debug("managed component not provided, will skip closing", "managedComponent", componentName)
			continue // we allow missing components: other types of nodes, tests, etc.
		}

		log.Debug("closing", "managedComponent", componentName)
		err := component.Close()
		if err != nil {
			if closeError == nil {
				closeError = ErrNodeCloseFailed
			}
			closeError = fmt.Errorf("%w, err: %s", closeError, err.Error())
		}
		log.Debug("closed", "managedComponent", componentName, "error", err)
	}

	return closeError
}

func find(needle string, haystack []string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}

	return false
}
