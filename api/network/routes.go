package network

import (
	"net/http"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
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

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	GetTotalStakedValue() (*api.StakeValues, error)
	GetDirectStakedList() ([]*api.DirectStakedValue, error)
	GetDelegatorsList() ([]*api.Delegator, error)
	StatusMetrics() external.StatusMetricsHandler
	GetAllIssuedESDTs(tokenType string) ([]string, error)
	IsInterfaceNil() bool
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, getConfigPath, GetNetworkConfig)
	router.RegisterHandler(http.MethodGet, getStatusPath, GetNetworkStatus)
	router.RegisterHandler(http.MethodGet, economicsPath, EconomicsMetrics)
	router.RegisterHandler(http.MethodGet, enableEpochsPath, GetEnableEpochs)
	router.RegisterHandler(http.MethodGet, getESDTsPath, getHandlerFuncForEsdt(""))
	router.RegisterHandler(http.MethodGet, getFFTsPath, getHandlerFuncForEsdt(core.FungibleESDT))
	router.RegisterHandler(http.MethodGet, getSFTsPath, getHandlerFuncForEsdt(core.SemiFungibleESDT))
	router.RegisterHandler(http.MethodGet, getNFTsPath, getHandlerFuncForEsdt(core.NonFungibleESDT))
	router.RegisterHandler(http.MethodGet, directStakedInfoPath, DirectStakedInfo)
	router.RegisterHandler(http.MethodGet, delegatedInfoPath, DelegatedInfo)
}

func getFacade(c *gin.Context) (FacadeHandler, bool) {
	facadeObj, ok := c.Get("facade")
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNilAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	facade, ok := facadeObj.(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	return facade, true
}

// GetNetworkConfig returns metrics related to the network configuration (shard independent)
func GetNetworkConfig(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	configMetrics := facade.StatusMetrics().ConfigMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"config": configMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetEnableEpochs returns metrics related to the activation epochs of the network
func GetEnableEpochs(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	enableEpochsMetrics := facade.StatusMetrics().EnableEpochsMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"enableEpochs": enableEpochsMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetNetworkStatus returns metrics related to the network status (shard specific)
func GetNetworkStatus(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	networkMetrics := facade.StatusMetrics().NetworkMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"status": networkMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// EconomicsMetrics is the endpoint that will return the economics data such as total supply
func EconomicsMetrics(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	stakeValues, err := facade.GetTotalStakedValue()
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

	metrics := facade.StatusMetrics().EconomicsMetrics()
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

func getHandlerFuncForEsdt(tokenType string) func(c *gin.Context) {
	return func(c *gin.Context) {
		facade, ok := getFacade(c)
		if !ok {
			return
		}

		tokens, err := facade.GetAllIssuedESDTs(tokenType)
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

// DirectStakedInfo is the endpoint that will return the directed staked info list
func DirectStakedInfo(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	directStakedList, err := facade.GetDirectStakedList()
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

// DelegatedInfo is the endpoint that will return the delegated list
func DelegatedInfo(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	delegatedList, err := facade.GetDelegatorsList()
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
