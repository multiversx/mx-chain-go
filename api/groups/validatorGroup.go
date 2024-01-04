package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

const (
	statisticsPath = "/statistics"
	auctionPath    = "/auction"
)

// validatorFacadeHandler defines the methods to be implemented by a facade for validator requests
type validatorFacadeHandler interface {
	ValidatorStatisticsApi() (map[string]*validator.ValidatorStatistics, error)
	AuctionListApi() ([]*common.AuctionListValidatorAPIResponse, error)
	IsInterfaceNil() bool
}

type validatorGroup struct {
	*baseGroup
	facade    validatorFacadeHandler
	mutFacade sync.RWMutex
}

// NewValidatorGroup returns a new instance of validatorGroup
func NewValidatorGroup(facade validatorFacadeHandler) (*validatorGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for validator group", errors.ErrNilFacadeHandler)
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
		{
			Path:    auctionPath,
			Method:  http.MethodGet,
			Handler: ng.auction,
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

// auction will return the list of the validators in the auction list
func (vg *validatorGroup) auction(c *gin.Context) {
	valStats, err := vg.getFacade().AuctionListApi()
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
			Data:  gin.H{"auctionList": valStats},
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
	castFacade, ok := newFacade.(validatorFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	vg.mutFacade.Lock()
	vg.facade = castFacade
	vg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vg *validatorGroup) IsInterfaceNil() bool {
	return vg == nil
}
