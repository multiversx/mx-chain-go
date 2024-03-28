package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/api/logs"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = logger.GetOrCreate("seednode/api")

// Start will boot up the api and appropriate routes, handlers and validators
func Start(restApiInterface string, marshalizer marshal.Marshalizer, p2pPrometheusMetricsEnabled bool) error {
	ws := gin.Default()
	ws.Use(cors.Default())

	registerRoutes(ws, marshalizer, p2pPrometheusMetricsEnabled)

	return ws.Run(restApiInterface)
}

func registerRoutes(ws *gin.Engine, marshalizer marshal.Marshalizer, p2pPrometheusMetricsEnabled bool) {
	registerLoggerWsRoute(ws, marshalizer, p2pPrometheusMetricsEnabled)
}

func registerLoggerWsRoute(ws *gin.Engine, marshalizer marshal.Marshalizer, p2pPrometheusMetricsEnabled bool) {
	upgrader := websocket.Upgrader{}

	ws.GET("/log", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls, err := logs.NewLogSender(marshalizer, conn, log)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls.StartSendingBlocking()
	})

	if p2pPrometheusMetricsEnabled {
		ws.GET("/debug/metrics/prometheus", gin.WrapH(promhttp.Handler()))
	}
}
