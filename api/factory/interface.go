package factory

import "github.com/gin-gonic/gin"

// HttpServerClosingHandler defines the basic actions of starting and closing that a web server should be able to do
type HttpServerClosingHandler interface {
	Start()
	Close() error
	IsInterfaceNil() bool
}

// MiddlewareProcessor defines a processor used internally by the web server when processing requests
type MiddlewareProcessor interface {
	MiddlewareHandlerFunc() gin.HandlerFunc
	IsInterfaceNil() bool
}

// MainApiHandler interface defines methods that can be used from `elrondFacade` context variable
type MainApiHandler interface {
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	IsInterfaceNil() bool
}

// HttpServerHandler defines the actions that a http server need to do
type HttpServerHandler interface {
	CreateHttpServer() (HttpServerClosingHandler, error)
	UpdateFacade(facade MainApiHandler) (HttpServerClosingHandler, error)
	IsInterfaceNil() bool
}
