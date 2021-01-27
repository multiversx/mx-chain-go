package network

import (
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/gin-gonic/gin"
)

const (
	getConfigPath   = "/config"
	getStatusPath   = "/status"
	economicsPath   = "/economics"
	totalStakedPath = "/total-staked"
)

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	GetTotalStakedValue() (*big.Int, error)
	StatusMetrics() external.StatusMetricsHandler
	IsInterfaceNil() bool
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, getConfigPath, GetNetworkConfig)
	router.RegisterHandler(http.MethodGet, getStatusPath, GetNetworkStatus)
	router.RegisterHandler(http.MethodGet, economicsPath, EconomicsMetrics)
	router.RegisterHandler(http.MethodGet, totalStakedPath, GetTotalStaked)
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

	metrics := facade.StatusMetrics().EconomicsMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": metrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetTotalStaked is the endpoint that will return the total staked value
func GetTotalStaked(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	totalStakedValue, err := facade.GetTotalStakedValue()
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

	totalStakedValueStr := "0"
	if totalStakedValue != nil {
		totalStakedValueStr = totalStakedValue.String()
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"totalStakedValue": totalStakedValueStr},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
