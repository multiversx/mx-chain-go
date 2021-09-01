package groups

import (
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api/groups")

type endpointProperties struct {
	isOpen bool
}

type baseGroup struct {
	endpoints []*shared.EndpointHandlerData
}

// GetEndpoints returns all the endpoints specific to the group
func (bg *baseGroup) GetEndpoints() []*shared.EndpointHandlerData {
	return bg.endpoints
}

// RegisterRoutes will register all the endpoints to the given web server
func (bg *baseGroup) RegisterRoutes(
	ws *gin.RouterGroup,
	apiConfig config.ApiRoutesConfig,
	additionalMiddlewares []shared.MiddlewareProcessor,
) {
	for _, handlerData := range bg.endpoints {
		properties := getEndpointProperties(ws, handlerData.Path, apiConfig)

		if !properties.isOpen {
			log.Debug("endpoint is closed", "path", handlerData.Path)
			continue
		}

		middlewares := make([]gin.HandlerFunc, 0)
		for _, middleware := range additionalMiddlewares {
			if middleware == nil ||
				middleware.MiddlewareHandlerFunc() == nil ||
				middleware.IsInterfaceNil() {
				continue
			}
			middlewares = append(middlewares, middleware.MiddlewareHandlerFunc())
		}

		beforeSpecifiesMiddlewares, afterSpecificMiddlewares := extractSpecificMiddlewares(handlerData.AdditionalMiddlewares)
		for _, specificMiddleware := range beforeSpecifiesMiddlewares {
			middlewares = append(middlewares, specificMiddleware)
		}
		middlewares = append(middlewares, handlerData.Handler)
		for _, specificMiddleware := range afterSpecificMiddlewares {
			middlewares = append(middlewares, specificMiddleware)
		}

		ws.Handle(handlerData.Method, handlerData.Path, middlewares...)
	}
}

func extractSpecificMiddlewares(middlewares []shared.AdditionalMiddleware) ([]gin.HandlerFunc, []gin.HandlerFunc) {
	if len(middlewares) == 0 {
		return nil, nil
	}

	beforeMiddlewares := make([]gin.HandlerFunc, 0)
	afterMiddlewares := make([]gin.HandlerFunc, 0)
	for _, middleware := range middlewares {
		if middleware.Before {
			beforeMiddlewares = append(beforeMiddlewares, middleware.Middleware)
		} else {
			afterMiddlewares = append(afterMiddlewares, middleware.Middleware)
		}
	}

	return beforeMiddlewares, afterMiddlewares
}

func getEndpointProperties(ws *gin.RouterGroup, path string, apiConfig config.ApiRoutesConfig) endpointProperties {
	basePath := ws.BasePath()

	// ws.BasePath will return paths like /group or /v1.0/group so we need the last token after splitting by /
	splitPath := strings.Split(basePath, "/")
	basePath = splitPath[len(splitPath)-1]

	group, ok := apiConfig.APIPackages[basePath]
	if !ok {
		return endpointProperties{
			isOpen: false,
		}
	}

	for _, route := range group.Routes {
		if route.Name == path {
			return endpointProperties{
				isOpen: route.Open,
			}
		}
	}

	return endpointProperties{
		isOpen: false,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bg *baseGroup) IsInterfaceNil() bool {
	return bg == nil
}
