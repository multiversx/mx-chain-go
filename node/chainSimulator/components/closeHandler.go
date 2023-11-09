package components

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

var errClose = errors.New("error while closing inner components")

type errorlessCloser interface {
	Close()
}

type allCloser interface {
	CloseAll() error
}

type closeHandler struct {
	mut        sync.RWMutex
	components []interface{}
}

// NewCloseHandler create a new closeHandler instance
func NewCloseHandler() *closeHandler {
	return &closeHandler{
		components: make([]interface{}, 0),
	}
}

// AddComponent will try to add a component to the inner list if that component is not nil
func (handler *closeHandler) AddComponent(component interface{}) {
	if check.IfNilReflect(component) {
		log.Error("programming error in closeHandler.AddComponent: nil component", "stack", string(debug.Stack()))
		return
	}

	handler.mut.Lock()
	handler.components = append(handler.components, component)
	handler.mut.Unlock()
}

// Close will try to close all components, wrapping errors, if necessary
func (handler *closeHandler) Close() error {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	var errorStrings []string
	for _, component := range handler.components {
		var err error

		switch t := component.(type) {
		case errorlessCloser:
			t.Close()
		case io.Closer:
			err = t.Close()
		case allCloser:
			err = t.CloseAll()
		}

		if err != nil {
			errorStrings = append(errorStrings, fmt.Errorf("%w while closing the component of type %T", err, component).Error())
		}
	}

	return AggregateErrors(errorStrings)
}

// AggregateErrors can aggregate all provided error strings into a single error variable
func AggregateErrors(errorStrings []string) error {
	if len(errorStrings) == 0 {
		return nil
	}

	return fmt.Errorf("%w %s", errClose, strings.Join(errorStrings, ", "))
}
