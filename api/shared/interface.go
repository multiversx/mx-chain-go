package shared

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-gonic/gin"
)

// HttpServerCloser defines the basic actions of starting and closing that a web server should be able to do
type HttpServerCloser interface {
	Start()
	Close() error
	IsInterfaceNil() bool
}

// MiddlewareProcessor defines a processor used internally by the web server when processing requests
type MiddlewareProcessor interface {
	MiddlewareHandlerFunc() gin.HandlerFunc
	IsInterfaceNil() bool
}

// ApiFacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type ApiFacadeHandler interface {
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	IsInterfaceNil() bool
}

// UpgradeableHttpServerHandler defines the actions that an upgradeable http server need to do
type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade ApiFacadeHandler) error
	Close() error
	IsInterfaceNil() bool
}

// GroupHandler defines the actions needed to be performed by an gin API group
type GroupHandler interface {
	UpdateFacade(newFacade interface{}) error
	RegisterRoutes(
		ws *gin.RouterGroup,
		apiConfig config.ApiRoutesConfig,
		additionalMiddlewares []MiddlewareProcessor,
	)
	IsInterfaceNil() bool
}
