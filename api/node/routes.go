package node

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/gin-gonic/gin"
)

const (
	pidQueryParam       = "pid"
	debugPath           = "/debug"
	heartbeatStatusPath = "/heartbeatstatus"
	metricsPath         = "/metrics"
	p2pStatusPath       = "/p2pstatus"
	peerInfoPath        = "/peerinfo"
	statusPath          = "/status"
)

// AccStateCheckpointsKey is used as a key for the number of account state checkpoints in the api response
const AccStateCheckpointsKey = "erd_num_accounts_state_checkpoints"

// PeerStateCheckpointsKey is used as a key for the number of peer state checkpoints in the api response
const PeerStateCheckpointsKey = "erd_num_peer_state_checkpoints"

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	GetHeartbeats() ([]data.PubKeyHeartbeat, error)
	StatusMetrics() external.StatusMetricsHandler
	GetQueryHandler(name string) (debug.QueryHandler, error)
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)
	GetNumCheckpointsFromAccountState() uint32
	GetNumCheckpointsFromPeerState() uint32
	IsInterfaceNil() bool
}

// QueryDebugRequest represents the structure on which user input for querying a debug info will validate against
type QueryDebugRequest struct {
	Name   string `form:"name" json:"name"`
	Search string `form:"search" json:"search"`
}

// Routes defines node related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, heartbeatStatusPath, HeartbeatStatus)
	router.RegisterHandler(http.MethodGet, statusPath, StatusMetrics)
	router.RegisterHandler(http.MethodGet, p2pStatusPath, P2pStatusMetrics)
	router.RegisterHandler(http.MethodGet, metricsPath, PrometheusMetrics)
	router.RegisterHandler(http.MethodPost, debugPath, QueryDebug)
	router.RegisterHandler(http.MethodGet, peerInfoPath, PeerInfo)
	// placeholder for custom routes
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

// HeartbeatStatus respond with the heartbeat status of the node
func HeartbeatStatus(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	hbStatus, err := facade.GetHeartbeats()
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

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"heartbeats": hbStatus},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// StatusMetrics returns the node statistics exported by an StatusMetricsHandler without p2p statistics
func StatusMetrics(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	details := facade.StatusMetrics().StatusMetricsMapWithoutP2P()
	details[AccStateCheckpointsKey] = facade.GetNumCheckpointsFromAccountState()
	details[PeerStateCheckpointsKey] = facade.GetNumCheckpointsFromPeerState()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": details},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// P2pStatusMetrics returns the node's p2p statistics exported by a StatusMetricsHandler
func P2pStatusMetrics(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	details := facade.StatusMetrics().StatusP2pMetricsMap()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": details},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// QueryDebug returns the debug information after the query has been interpreted
func QueryDebug(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	var gtx = QueryDebugRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	qh, err := facade.GetQueryHandler(gtx.Name)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrQueryError.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"result": qh.Query(gtx.Search)},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// PeerInfo returns the information of a provided p2p peer ID
func PeerInfo(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	queryVals := c.Request.URL.Query()
	pids := queryVals[pidQueryParam]
	pid := ""
	if len(pids) > 0 {
		pid = pids[0]
	}

	info, err := facade.GetPeerInfo(pid)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetPidInfo.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"info": info},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// PrometheusMetrics is the endpoint which will return the data in the way that prometheus expects them
func PrometheusMetrics(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	metrics := facade.StatusMetrics().StatusMetricsWithoutP2PPrometheusString()
	c.String(
		http.StatusOK,
		metrics,
	)
}
