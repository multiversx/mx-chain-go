package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
)

const (
	execManualTrigger    = "executed, trigger is affecting only the current node"
	execBroadcastTrigger = "executed, trigger is affecting current node and will get broadcast to other peers"
	triggerPath          = "/trigger"
)

// hardforkFacadeHandler defines the methods to be implemented by a facade for handling hardfork requests
type hardforkFacadeHandler interface {
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}

type hardforkGroup struct {
	*baseGroup
	facade    hardforkFacadeHandler
	mutFacade sync.RWMutex
}

// NewHardforkGroup returns a new instance of hardforkGroup
func NewHardforkGroup(facade hardforkFacadeHandler) (*hardforkGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for hardfork group", errors.ErrNilFacadeHandler)
	}

	hg := &hardforkGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    triggerPath,
			Method:  http.MethodPost,
			Handler: hg.triggerHandler,
		},
	}
	hg.endpoints = endpoints

	return hg, nil
}

// HardforkRequest represents the structure on which user input for triggering a hardfork will validate against
type HardforkRequest struct {
	Epoch               uint32 `form:"epoch" json:"epoch"`
	WithEarlyEndOfEpoch bool   `form:"withEarlyEndOfEpoch" json:"withEarlyEndOfEpoch"`
}

// triggerHandler will receive a trigger request from the client and propagate it for processing
func (hg *hardforkGroup) triggerHandler(c *gin.Context) {
	var hr = HardforkRequest{}
	err := c.ShouldBindJSON(&hr)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	err = hg.getFacade().Trigger(hr.Epoch, hr.WithEarlyEndOfEpoch)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	status := execManualTrigger
	if hg.getFacade().IsSelfTrigger() {
		status = execBroadcastTrigger
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"status": status},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (hg *hardforkGroup) getFacade() hardforkFacadeHandler {
	hg.mutFacade.RLock()
	defer hg.mutFacade.RUnlock()

	return hg.facade
}

// UpdateFacade will update the facade
func (hg *hardforkGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(hardforkFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	hg.mutFacade.Lock()
	hg.facade = castFacade
	hg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hg *hardforkGroup) IsInterfaceNil() bool {
	return hg == nil
}
