package groups

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
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
) {
	for _, handlerData := range bg.endpoints {
		properties := getEndpointProperties(ws, handlerData.Path, apiConfig)

		if !properties.isOpen {
			log.Debug("endpoint is closed", "path", handlerData.Path)
			continue
		}

		middlewares := make([]gin.HandlerFunc, 0)
		beforeSpecifiesMiddlewares, afterSpecificMiddlewares := extractSpecificMiddlewares(handlerData.AdditionalMiddlewares)

		middlewares = append(middlewares, beforeSpecifiesMiddlewares...)
		middlewares = append(middlewares, handlerData.Handler)
		middlewares = append(middlewares, afterSpecificMiddlewares...)

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
		if middleware.Position == shared.After {
			afterMiddlewares = append(afterMiddlewares, middleware.Middleware)
		} else {
			beforeMiddlewares = append(beforeMiddlewares, middleware.Middleware)
		}
	}

	return beforeMiddlewares, afterMiddlewares
}

func getEndpointProperties(ws *gin.RouterGroup, path string, apiConfig config.ApiRoutesConfig) endpointProperties {
	basePath := ws.BasePath()

	// ws.BasePath will return paths like /group or /v1.0/group, so we need the last token after splitting by /
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
