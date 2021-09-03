package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
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

	// AccStateCheckpointsKey is used as a key for the number of account state checkpoints in the api response
	AccStateCheckpointsKey = "erd_num_accounts_state_checkpoints"

	// PeerStateCheckpointsKey is used as a key for the number of peer state checkpoints in the api response
	PeerStateCheckpointsKey = "erd_num_peer_state_checkpoints"
)

type nodeFacadeHandler interface {
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

type nodeGroup struct {
	*baseGroup
	facade    nodeFacadeHandler
	mutFacade sync.RWMutex
}

// NewNodeGroup returns a new instance of nodeGroup
func NewNodeGroup(facade nodeFacadeHandler) (*nodeGroup, error) {
	if facade == nil {
		return nil, fmt.Errorf("%w for node group", errors.ErrNilFacadeHandler)
	}

	ng := &nodeGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    heartbeatStatusPath,
			Method:  http.MethodGet,
			Handler: ng.heartbeatStatus,
		},
		{
			Path:    statusPath,
			Method:  http.MethodGet,
			Handler: ng.statusMetrics,
		},
		{
			Path:    p2pStatusPath,
			Method:  http.MethodGet,
			Handler: ng.p2pStatusMetrics,
		},
		{
			Path:    metricsPath,
			Method:  http.MethodGet,
			Handler: ng.prometheusMetrics,
		},
		{
			Path:    debugPath,
			Method:  http.MethodPost,
			Handler: ng.queryDebug,
		},
		{
			Path:    peerInfoPath,
			Method:  http.MethodGet,
			Handler: ng.peerInfo,
		},

	}
	ng.endpoints = endpoints

	return ng, nil
}

// heartbeatStatus respond with the heartbeat status of the node
func (ng *nodeGroup) heartbeatStatus(c *gin.Context) {
	hbStatus, err := ng.getFacade().GetHeartbeats()
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

// statusMetrics returns the node statistics exported by an StatusMetricsHandler without p2p statistics
func (ng *nodeGroup) statusMetrics(c *gin.Context) {
	nodeFacade := ng.getFacade()
	details := nodeFacade.StatusMetrics().StatusMetricsMapWithoutP2P()
	details[AccStateCheckpointsKey] = nodeFacade.GetNumCheckpointsFromAccountState()
	details[PeerStateCheckpointsKey] = nodeFacade.GetNumCheckpointsFromPeerState()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": details},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// p2pStatusMetrics returns the node's p2p statistics exported by a StatusMetricsHandler
func (ng *nodeGroup) p2pStatusMetrics(c *gin.Context) {
	details := ng.getFacade().StatusMetrics().StatusP2pMetricsMap()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"metrics": details},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// queryDebug returns the debug information after the query has been interpreted
func (ng *nodeGroup) queryDebug(c *gin.Context) {
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

	qh, err := ng.getFacade().GetQueryHandler(gtx.Name)
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

// peerInfo returns the information of a provided p2p peer ID
func (ng *nodeGroup) peerInfo(c *gin.Context) {
	queryVals := c.Request.URL.Query()
	pids := queryVals[pidQueryParam]
	pid := ""
	if len(pids) > 0 {
		pid = pids[0]
	}

	info, err := ng.getFacade().GetPeerInfo(pid)
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

// prometheusMetrics is the endpoint which will return the data in the way that prometheus expects them
func (ng *nodeGroup) prometheusMetrics(c *gin.Context) {
	metrics := ng.getFacade().StatusMetrics().StatusMetricsWithoutP2PPrometheusString()
	c.String(
		http.StatusOK,
		metrics,
	)
}


func (ng *nodeGroup) getFacade() nodeFacadeHandler {
	ng.mutFacade.RLock()
	defer ng.mutFacade.RUnlock()

	return ng.facade
}

// UpdateFacade will update the facade
func (ng *nodeGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(nodeFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	ng.mutFacade.Lock()
	ng.facade = castFacade
	ng.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ng *nodeGroup) IsInterfaceNil() bool {
	return ng == nil
}
