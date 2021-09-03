package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/gin-gonic/gin"
)

const (
	getConfigPath        = "/config"
	getStatusPath        = "/status"
	economicsPath        = "/economics"
	enableEpochsPath     = "/enable-epochs"
	getESDTsPath         = "/esdts"
	getFFTsPath          = "/esdt/fungible-tokens"
	getSFTsPath          = "/esdt/semi-fungible-tokens"
	getNFTsPath          = "/esdt/non-fungible-tokens"
	directStakedInfoPath = "/direct-staked-info"
	delegatedInfoPath    = "/delegated-info"
)

type networkFacadeHandler interface {
	GetTotalStakedValue() (*api.StakeValues, error)
	GetDirectStakedList() ([]*api.DirectStakedValue, error)
	GetDelegatorsList() ([]*api.Delegator, error)
	StatusMetrics() external.StatusMetricsHandler
	GetAllIssuedESDTs(tokenType string) ([]string, error)
	IsInterfaceNil() bool
}

type networkGroup struct {
	*baseGroup
	facade    networkFacadeHandler
	mutFacade sync.RWMutex
}

// NewNetworkGroup returns a new instance of networkGroup
func NewNetworkGroup(facade networkFacadeHandler) (*networkGroup, error) {
	if facade == nil {
		return nil, fmt.Errorf("%w for network group", errors.ErrNilFacadeHandler)
	}

	ng := &networkGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getConfigPath,
			Method:  http.MethodGet,
			Handler: ng.getNetworkConfig,
		},
		{
			Path:    getStatusPath,
			Method:  http.MethodGet,
			Handler: ng.getNetworkStatus,
		},
		{
			Path: economicsPath,
			Method: http.MethodGet,
			Handler: ng.economicsMetrics,
		},
		{
			Path: enableEpochsPath,
			Method: http.MethodGet,
			Handler: ng.getEnableEpochs,
		},
		{
			Path: getESDTsPath,
			Method: http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(""),
		},
		{
			Path: getFFTsPath,
			Method: http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(core.FungibleESDT),
		},
		{
			Path: getSFTsPath,
			Method: http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(core.SemiFungibleESDT),
		},
		{
			Path: getNFTsPath,
			Method: http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(core.NonFungibleESDT),
		},
		{
			Path: directStakedInfoPath,
			Method: http.MethodGet,
			Handler: ng.directStakedInfo,
		},
		{
			Path: delegatedInfoPath,
			Method: http.MethodGet,
			Handler: ng.delegatedInfo,
		},
	}
	ng.endpoints = endpoints

	return ng, nil
}

// getNetworkConfig returns metrics related to the network configuration (shard independent)
func (ng *networkGroup) getNetworkConfig(c *gin.Context) {
	configMetrics := ng.getFacade().StatusMetrics().ConfigMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"config": configMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getEnableEpochs returns metrics related to the activation epochs of the network
func (ng *networkGroup) getEnableEpochs(c *gin.Context) {
	enableEpochsMetrics := ng.getFacade().StatusMetrics().EnableEpochsMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"enableEpochs": enableEpochsMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getNetworkStatus returns metrics related to the network status (shard specific)
func (ng *networkGroup) getNetworkStatus(c *gin.Context) {
	networkMetrics := ng.getFacade().StatusMetrics().NetworkMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"status": networkMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// economicsMetrics is the endpoint that will return the economics data such as total supply
func (ng *networkGroup) economicsMetrics(c *gin.Context) {
	stakeValues, err := ng.getFacade().GetTotalStakedValue()
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

	metrics := ng.getFacade().StatusMetrics().EconomicsMetrics()
	metrics[common.MetricTotalBaseStakedValue] = stakeValues.BaseStaked.String()
	metrics[common.MetricTopUpValue] = stakeValues.TopUp.String()

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": metrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (ng *networkGroup) getHandlerFuncForEsdt(tokenType string) func(c *gin.Context) {
	return func(c *gin.Context) {
		tokens, err := ng.getFacade().GetAllIssuedESDTs(tokenType)
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

		c.JSON(
			http.StatusOK,
			shared.GenericAPIResponse{
				Data:  gin.H{"tokens": tokens},
				Error: "",
				Code:  shared.ReturnCodeSuccess,
			},
		)
	}
}

// directStakedInfo is the endpoint that will return the directed staked info list
func (ng *networkGroup) directStakedInfo(c *gin.Context) {
	directStakedList, err := ng.getFacade().GetDirectStakedList()
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

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"list": directStakedList},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// delegatedInfo is the endpoint that will return the delegated list
func (ng *networkGroup) delegatedInfo(c *gin.Context) {
	delegatedList, err := ng.getFacade().GetDelegatorsList()
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

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"list": delegatedList},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (ng *networkGroup) getFacade() networkFacadeHandler {
	ng.mutFacade.RLock()
	defer ng.mutFacade.RUnlock()

	return ng.facade
}

// UpdateFacade will update the facade
func (ng *networkGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(networkFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	ng.mutFacade.Lock()
	ng.facade = castFacade
	ng.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ng *networkGroup) IsInterfaceNil() bool {
	return ng == nil
}
