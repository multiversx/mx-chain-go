package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	execManualTrigger    = "executed, trigger is affecting only the current node"
	execBroadcastTrigger = "executed, trigger is affecting current node and will get broadcast to other peers"
	triggerPath          = "/trigger"
)

// HardforkFacadeHandler defines the methods to be implemented by a facade for handling hardfork requests
type HardforkFacadeHandler interface {
	Trigger(epoch uint32, withEarlyEndOfEpoch bool) error
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}

type hardforkGroup struct {
	facade    HardforkFacadeHandler
	mutFacade sync.RWMutex
	*baseGroup
}

// NewHardforkGroup returns a new instance of hardforkGroup
func NewHardforkGroup(facadeHandler interface{}) (*hardforkGroup, error) {
	if facadeHandler == nil {
		return nil, errors.ErrNilFacadeHandler
	}

	facade, ok := facadeHandler.(HardforkFacadeHandler)
	if !ok {
		return nil, fmt.Errorf("%w for hardfork group", errors.ErrFacadeWrongTypeAssertion)
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

// HarforkRequest represents the structure on which user input for triggering a hardfork will validate against
type HarforkRequest struct {
	Epoch               uint32 `form:"epoch" json:"epoch"`
	WithEarlyEndOfEpoch bool   `form:"withEarlyEndOfEpoch" json:"withEarlyEndOfEpoch"`
}

// triggerHandler will receive a trigger request from the client and propagate it for processing
func (hg *hardforkGroup) triggerHandler(c *gin.Context) {
	var hr = HarforkRequest{}
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

func (hg *hardforkGroup) getFacade() HardforkFacadeHandler {
	hg.mutFacade.RLock()
	defer hg.mutFacade.RUnlock()

	return hg.facade
}

// UpdateFacade will update the facade
func (hg *hardforkGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castedFacade, ok := newFacade.(HardforkFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	hg.mutFacade.Lock()
	hg.facade = castedFacade
	hg.mutFacade.Unlock()

	return nil
}
