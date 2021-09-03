package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/gin-gonic/gin"
)

const statisticsPath = "/statistics"

// validatorFacadeHandler defines the methods to be implemented by a facade for validator requests
type validatorFacadeHandler interface {
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	IsInterfaceNil() bool
}

type validatorGroup struct {
	facade    validatorFacadeHandler
	mutFacade sync.RWMutex
	*baseGroup
}

// NewValidatorGroup returns a new instance of validatorGroup
func NewValidatorGroup(facadeHandler interface{}) (*validatorGroup, error) {
	if facadeHandler == nil {
		return nil, errors.ErrNilFacadeHandler
	}

	facade, ok := facadeHandler.(validatorFacadeHandler)
	if !ok {
		return nil, fmt.Errorf("%w for validator group", errors.ErrFacadeWrongTypeAssertion)
	}

	ng := &validatorGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    statisticsPath,
			Method:  http.MethodGet,
			Handler: ng.statistics,
		},
	}
	ng.endpoints = endpoints

	return ng, nil
}

// statistics will return the validation statistics for all validators
func (vg *validatorGroup) statistics(c *gin.Context) {
	valStats, err := vg.getFacade().ValidatorStatisticsApi()
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"statistics": valStats},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (vg *validatorGroup) getFacade() validatorFacadeHandler {
	vg.mutFacade.RLock()
	defer vg.mutFacade.RUnlock()

	return vg.facade
}

// UpdateFacade will update the facade
func (vg *validatorGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castedFacade, ok := newFacade.(validatorFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	vg.mutFacade.Lock()
	vg.facade = castedFacade
	vg.mutFacade.Unlock()

	return nil
}
