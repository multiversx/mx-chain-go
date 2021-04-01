package network

import (
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/gin-gonic/gin"
)

const (
	getConfigPath    = "/config"
	getStatusPath    = "/status"
	economicsPath    = "/economics"
	enableEpochsPath = "/enable-epochs"
)

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	GetTotalStakedValue() (*api.StakeValues, error)
	StatusMetrics() external.StatusMetricsHandler
	IsInterfaceNil() bool
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, getConfigPath, GetNetworkConfig)
	router.RegisterHandler(http.MethodGet, getStatusPath, GetNetworkStatus)
	router.RegisterHandler(http.MethodGet, economicsPath, EconomicsMetrics)
	router.RegisterHandler(http.MethodGet, enableEpochsPath, GetEnableEpochs)
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

	enableEpochsMetrics := facade.StatusMetrics().EnableEpochMetrics()
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
	metrics[core.MetricTotalStakedValue] = stakeValues.TotalStaked.String()
	metrics[core.MetricTopUpValue] = stakeValues.TopUp.String()

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": metrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
