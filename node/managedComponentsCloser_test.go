package node

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
)

func TestNewManagedComponentsCloserShouldWork(t *testing.T) {
	t.Parallel()

	closerHandler := newManagedComponentsCloser()
	assert.NotNil(t, closerHandler)
}

func TestManagedComponentsCloser_AddManagedComponent(t *testing.T) {
	t.Parallel()

	closerHandler := newManagedComponentsCloser()
	t.Run("nil component should error", func(t *testing.T) {
		err := closerHandler.addManagedComponent(nil)
		assert.ErrorIs(t, err, ErrNilComponentHandler)
		assert.Equal(t, 0, len(closerHandler.allComponents))
	})
	t.Run("component name is unknown should error", func(t *testing.T) {
		testComponent := components.NewComponentHandlerStub("unknown-component")

		err := closerHandler.addManagedComponent(testComponent)
		assert.ErrorIs(t, err, ErrInvalidComponentHandler)
		assert.Contains(t, err.Error(), "identifier: unknown-component")
		assert.Equal(t, 0, len(closerHandler.allComponents))
	})
	t.Run("double adding the same component should error", func(t *testing.T) {
		testComponent := components.NewComponentHandlerStub(factory.BootstrapComponentsName)

		err := closerHandler.addManagedComponent(testComponent)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(closerHandler.allComponents))

		err = closerHandler.addManagedComponent(testComponent)
		assert.ErrorIs(t, err, ErrDuplicatedComponentHandler)
		assert.Equal(t, 1, len(closerHandler.allComponents))
	})
}

func TestManagedComponentsCloser_CloseAllKnownComponentsShouldWork(t *testing.T) {
	t.Parallel()

	counter := 0
	closingOrderMap := make(map[string]int)

	closerHandler := newManagedComponentsCloser()

	// deliberately mixed the order of components added

	processComponent := components.NewComponentHandlerStub(factory.ProcessComponentsName)
	processComponent.CloseCalled = createCloseHandler(processComponent, closingOrderMap, &counter)
	err := closerHandler.addManagedComponent(processComponent)
	assert.Nil(t, err)

	consensusComponent := components.NewComponentHandlerStub(factory.ConsensusComponentsName)
	consensusComponent.CloseCalled = createCloseHandler(consensusComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(consensusComponent)
	assert.Nil(t, err)

	heartbeatComponent := components.NewComponentHandlerStub(factory.HeartbeatV2ComponentsName)
	heartbeatComponent.CloseCalled = createCloseHandler(heartbeatComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(heartbeatComponent)
	assert.Nil(t, err)

	cryptoComponent := components.NewComponentHandlerStub(factory.CryptoComponentsName)
	cryptoComponent.CloseCalled = createCloseHandler(cryptoComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(cryptoComponent)
	assert.Nil(t, err)

	coreComponent := components.NewComponentHandlerStub(factory.CoreComponentsName)
	coreComponent.CloseCalled = createCloseHandler(coreComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(coreComponent)
	assert.Nil(t, err)

	statusCoreComponent := components.NewComponentHandlerStub(factory.StatusCoreComponentsName)
	statusCoreComponent.CloseCalled = createCloseHandler(statusCoreComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(statusCoreComponent)
	assert.Nil(t, err)

	networkComponent := components.NewComponentHandlerStub(factory.NetworkComponentsName)
	networkComponent.CloseCalled = createCloseHandler(networkComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(networkComponent)
	assert.Nil(t, err)

	statusComponent := components.NewComponentHandlerStub(factory.StatusComponentsName)
	statusComponent.CloseCalled = createCloseHandler(statusComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(statusComponent)
	assert.Nil(t, err)

	bootstrapComponent := components.NewComponentHandlerStub(factory.BootstrapComponentsName)
	bootstrapComponent.CloseCalled = createCloseHandler(bootstrapComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(bootstrapComponent)
	assert.Nil(t, err)

	stateComponent := components.NewComponentHandlerStub(factory.StateComponentsName)
	stateComponent.CloseCalled = createCloseHandler(stateComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(stateComponent)
	assert.Nil(t, err)

	dataComponent := components.NewComponentHandlerStub(factory.DataComponentsName)
	dataComponent.CloseCalled = createCloseHandler(dataComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(dataComponent)
	assert.Nil(t, err)

	err = closerHandler.close()
	assert.Nil(t, err)

	expectedOrderMap := map[string]int{
		factory.NetworkComponentsName:     0,
		factory.ConsensusComponentsName:   1,
		factory.HeartbeatV2ComponentsName: 2,
		factory.ProcessComponentsName:     3,
		factory.StatusComponentsName:      4,
		factory.StateComponentsName:       5,
		factory.DataComponentsName:        6,
		factory.BootstrapComponentsName:   7,
		factory.CryptoComponentsName:      8,
		factory.CoreComponentsName:        9,
		factory.StatusCoreComponentsName:  10,
	}

	assert.Equal(t, expectedOrderMap, closingOrderMap)
	assert.Equal(t, 11, len(getClosingOrderForManagedComponents())) // test that there are no other components
}

func TestManagedComponentsCloser_CloseOnlySomeComponentsShouldWork(t *testing.T) {
	t.Parallel()

	counter := 0
	closingOrderMap := make(map[string]int)

	closerHandler := newManagedComponentsCloser()

	// deliberately mixed the order of components added

	processComponent := components.NewComponentHandlerStub(factory.ProcessComponentsName)
	processComponent.CloseCalled = createCloseHandler(processComponent, closingOrderMap, &counter)
	err := closerHandler.addManagedComponent(processComponent)
	assert.Nil(t, err)

	coreComponent := components.NewComponentHandlerStub(factory.CoreComponentsName)
	coreComponent.CloseCalled = createCloseHandler(coreComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(coreComponent)
	assert.Nil(t, err)

	stateComponent := components.NewComponentHandlerStub(factory.StateComponentsName)
	stateComponent.CloseCalled = createCloseHandler(stateComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(stateComponent)
	assert.Nil(t, err)

	dataComponent := components.NewComponentHandlerStub(factory.DataComponentsName)
	dataComponent.CloseCalled = createCloseHandler(dataComponent, closingOrderMap, &counter)
	err = closerHandler.addManagedComponent(dataComponent)
	assert.Nil(t, err)

	err = closerHandler.close()
	assert.Nil(t, err)

	expectedOrderMap := map[string]int{
		factory.ProcessComponentsName: 0,
		factory.StateComponentsName:   1,
		factory.DataComponentsName:    2,
		factory.CoreComponentsName:    3,
	}

	assert.Equal(t, expectedOrderMap, closingOrderMap)
}

func TestManagedComponentsCloser_CloseAllComponentsEvenIfSomeError(t *testing.T) {
	t.Parallel()

	closerHandler := newManagedComponentsCloser()

	counter := 0
	closingOrderMap := make(map[string]int)

	errCloseProcess := errors.New("process close error")
	processComponent := components.NewComponentHandlerStub(factory.ProcessComponentsName)
	processComponent.CloseCalled = func() error {
		closingOrderMap[processComponent.Name] = counter
		counter++

		return errCloseProcess
	}
	err := closerHandler.addManagedComponent(processComponent)
	assert.Nil(t, err)

	coreComponent := components.NewComponentHandlerStub(factory.CoreComponentsName)
	coreComponent.CloseCalled = func() error {
		closingOrderMap[coreComponent.Name] = counter
		counter++

		return nil
	}
	err = closerHandler.addManagedComponent(coreComponent)
	assert.Nil(t, err)

	errCloseState := errors.New("state close error")
	stateComponent := components.NewComponentHandlerStub(factory.StateComponentsName)
	stateComponent.CloseCalled = func() error {
		closingOrderMap[stateComponent.Name] = counter
		counter++

		return errCloseState
	}
	err = closerHandler.addManagedComponent(stateComponent)
	assert.Nil(t, err)

	dataComponent := components.NewComponentHandlerStub(factory.DataComponentsName)
	dataComponent.CloseCalled = func() error {
		closingOrderMap[dataComponent.Name] = counter
		counter++

		return nil
	}
	err = closerHandler.addManagedComponent(dataComponent)
	assert.Nil(t, err)

	err = closerHandler.close()
	assert.ErrorIs(t, err, ErrNodeCloseFailed)
	assert.Contains(t, err.Error(), errCloseProcess.Error())
	assert.Contains(t, err.Error(), errCloseState.Error())

	expectedOrderMap := map[string]int{
		factory.ProcessComponentsName: 0,
		factory.StateComponentsName:   1,
		factory.DataComponentsName:    2,
		factory.CoreComponentsName:    3,
	}

	assert.Equal(t, expectedOrderMap, closingOrderMap)
}

func createCloseHandler(component factory.ComponentHandler, receivedMap map[string]int, counter *int) func() error {
	return func() error {
		receivedMap[component.String()] = *counter
		*counter++
		return nil
	}
}
