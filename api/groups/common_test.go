package groups_test

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func startWebServer(group shared.GroupHandler, path string, apiConfig config.ApiRoutesConfig) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	routes := ws.Group(path)
	group.RegisterRoutes(routes, apiConfig)
	return ws
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	logError(err)
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
