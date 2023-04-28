package api

import (
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/config"
)

// GroupHandlerStub -
type GroupHandlerStub struct {
	UpdateFacadeCalled   func(facade interface{}) error
	RegisterRoutesCalled func(ws *gin.RouterGroup, apiConfig config.ApiRoutesConfig)
}

// UpdateFacade -
func (stub *GroupHandlerStub) UpdateFacade(facade interface{}) error {
	if stub.UpdateFacadeCalled != nil {
		return stub.UpdateFacadeCalled(facade)
	}
	return nil
}

// RegisterRoutes -
func (stub *GroupHandlerStub) RegisterRoutes(ws *gin.RouterGroup, apiConfig config.ApiRoutesConfig) {
	if stub.RegisterRoutesCalled != nil {
		stub.RegisterRoutesCalled(ws, apiConfig)
	}
}

// IsInterfaceNil -
func (stub *GroupHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
