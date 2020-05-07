package wrapper

import (
	"errors"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-gonic/gin"
)

// RouterWrapper is a wrapper over the gin RouterGroup in order to handle the logic of enabling or disabling routes
type RouterWrapper struct {
	router          *gin.RouterGroup
	routesConfig    config.APIPackageConfig
	mutRoutesConfig sync.RWMutex
}

// NewRouterWrapper will return a new instance of RouterWrapper
func NewRouterWrapper(packageName string, router *gin.RouterGroup, routesConfig config.ApiRoutesConfig) (*RouterWrapper, error) {
	if router == nil {
		return nil, ErrNilRouter
	}

	configForPackage, ok := routesConfig.APIPackages[packageName]
	if !ok {
		return nil, errors.New("config not found")
	}

	return &RouterWrapper{
		router:       router,
		routesConfig: configForPackage,
	}, nil
}

// RegisterHandler will register the handler for the given method and path
func (rw *RouterWrapper) RegisterHandler(method string, path string, handlers ...gin.HandlerFunc) {
	if rw.isEndpointActive(path) {
		rw.router.Handle(method, path, handlers...)
	}
}

func (rw *RouterWrapper) isEndpointActive(endpointToCheck string) bool {
	rw.mutRoutesConfig.RLock()
	routesConfig := rw.routesConfig
	rw.mutRoutesConfig.RUnlock()
	for _, endpoint := range routesConfig.Routes {
		if endpoint.Name == endpointToCheck && endpoint.Open {
			return true
		}
	}

	return false
}
