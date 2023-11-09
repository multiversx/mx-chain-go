package groups

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/api/shared/logging"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/external"
)

const (
	getConfigPath          = "/config"
	getStatusPath          = "/status"
	economicsPath          = "/economics"
	enableEpochsPath       = "/enable-epochs"
	getESDTsPath           = "/esdts"
	getFFTsPath            = "/esdt/fungible-tokens"
	getSFTsPath            = "/esdt/semi-fungible-tokens"
	getNFTsPath            = "/esdt/non-fungible-tokens"
	getESDTSupplyPath      = "/esdt/supply/:token"
	directStakedInfoPath   = "/direct-staked-info"
	delegatedInfoPath      = "/delegated-info"
	ratingsPath            = "/ratings"
	genesisNodesConfigPath = "/genesis-nodes"
	genesisBalances        = "/genesis-balances"
	gasConfigPath          = "/gas-configs"
)

// networkFacadeHandler defines the methods to be implemented by a facade for handling network requests
type networkFacadeHandler interface {
	GetTotalStakedValue() (*api.StakeValues, error)
	GetDirectStakedList() ([]*api.DirectStakedValue, error)
	GetDelegatorsList() ([]*api.Delegator, error)
	StatusMetrics() external.StatusMetricsHandler
	GetAllIssuedESDTs(tokenType string) ([]string, error)
	GetTokenSupply(token string) (*api.ESDTSupply, error)
	GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string, error)
	GetGenesisBalances() ([]*common.InitialAccountAPI, error)
	GetGasConfigs() (map[string]map[string]uint64, error)
	IsInterfaceNil() bool
}

// GenesisNodesConfig defines the eligible and waiting nodes configurations
type GenesisNodesConfig struct {
	Eligible map[uint32][]string `json:"eligible"`
	Waiting  map[uint32][]string `json:"waiting"`
}

// GasConfig defines the gas config sections to be exposed
type GasConfig struct {
	BuiltInCost            map[string]uint64 `json:"builtInCost"`
	MetaChainSystemSCsCost map[string]uint64 `json:"metaSystemSCCost"`
}

type networkGroup struct {
	*baseGroup
	facade    networkFacadeHandler
	mutFacade sync.RWMutex
}

// NewNetworkGroup returns a new instance of networkGroup
func NewNetworkGroup(facade networkFacadeHandler) (*networkGroup, error) {
	if check.IfNil(facade) {
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
			Path:    economicsPath,
			Method:  http.MethodGet,
			Handler: ng.economicsMetrics,
		},
		{
			Path:    enableEpochsPath,
			Method:  http.MethodGet,
			Handler: ng.getEnableEpochs,
		},
		{
			Path:    getESDTsPath,
			Method:  http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(""),
		},
		{
			Path:    getFFTsPath,
			Method:  http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(core.FungibleESDT),
		},
		{
			Path:    getSFTsPath,
			Method:  http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(core.SemiFungibleESDT),
		},
		{
			Path:    getNFTsPath,
			Method:  http.MethodGet,
			Handler: ng.getHandlerFuncForEsdt(core.NonFungibleESDT),
		},
		{
			Path:    directStakedInfoPath,
			Method:  http.MethodGet,
			Handler: ng.directStakedInfo,
		},
		{
			Path:    delegatedInfoPath,
			Method:  http.MethodGet,
			Handler: ng.delegatedInfo,
		},
		{
			Path:    getESDTSupplyPath,
			Method:  http.MethodGet,
			Handler: ng.getESDTTokenSupply,
		},
		{
			Path:    ratingsPath,
			Method:  http.MethodGet,
			Handler: ng.getRatingsConfig,
		},
		{
			Path:    genesisNodesConfigPath,
			Method:  http.MethodGet,
			Handler: ng.getGenesisNodesConfig,
		},
		{
			Path:    genesisBalances,
			Method:  http.MethodGet,
			Handler: ng.getGenesisBalances,
		},
		{
			Path:    gasConfigPath,
			Method:  http.MethodGet,
			Handler: ng.getGasConfig,
		},
	}
	ng.endpoints = endpoints

	return ng, nil
}

// getNetworkConfig returns metrics related to the network configuration (shard independent)
func (ng *networkGroup) getNetworkConfig(c *gin.Context) {
	configMetrics, err := ng.getFacade().StatusMetrics().ConfigMetrics()
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
			Data:  gin.H{"config": configMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getEnableEpochs returns metrics related to the activation epochs of the network
func (ng *networkGroup) getEnableEpochs(c *gin.Context) {
	enableEpochsMetrics, err := ng.getFacade().StatusMetrics().EnableEpochsMetrics()
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
			Data:  gin.H{"enableEpochs": enableEpochsMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getNetworkStatus returns metrics related to the network status (shard specific)
func (ng *networkGroup) getNetworkStatus(c *gin.Context) {
	networkMetrics, err := ng.getFacade().StatusMetrics().NetworkMetrics()
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

	metrics, err := ng.getFacade().StatusMetrics().EconomicsMetrics()
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

func (ng *networkGroup) getESDTTokenSupply(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrBadUrlParams)
		return
	}

	supply, err := ng.getFacade().GetTokenSupply(token)
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
			Data:  supply,
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getRatingsConfig returns metrics related to ratings configuration
func (ng *networkGroup) getRatingsConfig(c *gin.Context) {
	ratingsConfig, err := ng.getFacade().StatusMetrics().RatingsMetrics()
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
			Data:  gin.H{"config": ratingsConfig},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getGenesisNodesConfig return genesis nodes configuration
func (ng *networkGroup) getGenesisNodesConfig(c *gin.Context) {
	start := time.Now()
	eligibleNodesConfig, waitingNodesConfig, err := ng.getFacade().GetGenesisNodesPubKeys()
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetGenesisNodesPubKeys")

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetGenesisNodes.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	nc := GenesisNodesConfig{
		Eligible: eligibleNodesConfig,
		Waiting:  waitingNodesConfig,
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"nodes": nc}, "", shared.ReturnCodeSuccess)
}

// getGenesisBalances return genesis balances configuration
func (ng *networkGroup) getGenesisBalances(c *gin.Context) {
	start := time.Now()
	genesisBalances, err := ng.getFacade().GetGenesisBalances()
	log.Debug("API call: GetGenesisBalances", "duration", time.Since(start))

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetGenesisBalances.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"balances": genesisBalances}, "", shared.ReturnCodeSuccess)
}

// getGasConfig returns currently used gas configs configuration
func (ng *networkGroup) getGasConfig(c *gin.Context) {
	gasConfigMap, err := ng.getFacade().GetGasConfigs()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetGasConfigs.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	gc := GasConfig{
		BuiltInCost:            gasConfigMap[common.BuiltInCost],
		MetaChainSystemSCsCost: gasConfigMap[common.MetaChainSystemSCsCost],
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"gasConfigs": gc}, "", shared.ReturnCodeSuccess)
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
		return fmt.Errorf("%w for network group", errors.ErrFacadeWrongTypeAssertion)
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
