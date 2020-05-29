package network

import (
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	StatusMetrics() external.StatusMetricsHandler
	IsInterfaceNil() bool
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, "/config", ConfigData)
	router.RegisterHandler(http.MethodGet, "/status", GetNetworkData)
}

// ConfigData returns data about current configuration
func ConfigData(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			core.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  string(core.ReturnCodeInternalError),
			},
		)
		return
	}

	configMetrics := ef.StatusMetrics().ConfigMetrics()
	c.JSON(
		http.StatusOK,
		core.GenericAPIResponse{
			Data:  gin.H{"config": configMetrics},
			Error: "",
			Code:  string(core.ReturnCodeSuccess),
		},
	)
}

// GetNetworkData returns data about current network metrics
func GetNetworkData(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			core.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  string(core.ReturnCodeInternalError),
			},
		)
		return
	}

	networkMetrics := ef.StatusMetrics().NetworkMetrics()
	c.JSON(
		http.StatusOK,
		core.GenericAPIResponse{
			Data:  gin.H{"status": networkMetrics},
			Error: "",
			Code:  string(core.ReturnCodeSuccess),
		},
	)
}
