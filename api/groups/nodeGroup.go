package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
)

const (
	pidQueryParam             = "pid"
	debugPath                 = "/debug"
	heartbeatStatusPath       = "/heartbeatstatus"
	metricsPath               = "/metrics"
	p2pStatusPath             = "/p2pstatus"
	peerInfoPath              = "/peerinfo"
	statusPath                = "/status"
	epochStartDataForEpoch    = "/epoch-start/:epoch"
	bootstrapStatusPath       = "/bootstrapstatus"
	connectedPeersRatingsPath = "/connected-peers-ratings"
	managedKeys               = "/managed-keys"
	managedKeysCount          = "/managed-keys/count"
	eligibleManagedKeys       = "/managed-keys/eligible"
	waitingManagedKeys        = "/managed-keys/waiting"
)

// nodeFacadeHandler defines the methods to be implemented by a facade for node requests
type nodeFacadeHandler interface {
	GetHeartbeats() ([]data.PubKeyHeartbeat, error)
	StatusMetrics() external.StatusMetricsHandler
	GetQueryHandler(name string) (debug.QueryHandler, error)
	GetEpochStartDataAPI(epoch uint32) (*common.EpochStartDataAPI, error)
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)
	GetConnectedPeersRatingsOnMainNetwork() (string, error)
	GetManagedKeysCount() int
	GetManagedKeys() []string
	GetEligibleManagedKeys() ([]string, error)
	GetWaitingManagedKeys() ([]string, error)
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
	if check.IfNil(facade) {
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
		{
			Path:    epochStartDataForEpoch,
			Method:  http.MethodGet,
			Handler: ng.epochStartDataForEpoch,
		},
		{
			Path:    bootstrapStatusPath,
			Method:  http.MethodGet,
			Handler: ng.bootstrapMetrics,
		},
		{
			Path:    connectedPeersRatingsPath,
			Method:  http.MethodGet,
			Handler: ng.connectedPeersRatings,
		},
		{
			Path:    managedKeysCount,
			Method:  http.MethodGet,
			Handler: ng.managedKeysCount,
		},
		{
			Path:    managedKeys,
			Method:  http.MethodGet,
			Handler: ng.managedKeys,
		},
		{
			Path:    eligibleManagedKeys,
			Method:  http.MethodGet,
			Handler: ng.managedKeysEligible,
		},
		{
			Path:    waitingManagedKeys,
			Method:  http.MethodGet,
			Handler: ng.managedKeysWaiting,
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
	metrics, err := nodeFacade.StatusMetrics().StatusMetricsMapWithoutP2P()
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
			Data:  gin.H{"metrics": metrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// p2pStatusMetrics returns the node's p2p statistics exported by a StatusMetricsHandler
func (ng *nodeGroup) p2pStatusMetrics(c *gin.Context) {
	metrics, err := ng.getFacade().StatusMetrics().StatusP2pMetricsMap()
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
			Data:  gin.H{"metrics": metrics},
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

// epochStartDataForEpoch returns epoch start data for the provided epoch
func (ng *nodeGroup) epochStartDataForEpoch(c *gin.Context) {
	epoch, err := getQueryParamEpoch(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrValidation, errors.ErrBadUrlParams)
		return
	}

	epochStartData, err := ng.getFacade().GetEpochStartDataAPI(epoch)
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetEpochStartData, err)
		return
	}

	shared.RespondWithSuccess(c, gin.H{"epochStart": epochStartData})
}

// prometheusMetrics is the endpoint which will return the data in the way that prometheus expects them
func (ng *nodeGroup) prometheusMetrics(c *gin.Context) {
	metrics, err := ng.getFacade().StatusMetrics().StatusMetricsWithoutP2PPrometheusString()
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

	c.String(
		http.StatusOK,
		metrics,
	)
}

// bootstrapMetrics returns the node's bootstrap statistics exported by a StatusMetricsHandler
func (ng *nodeGroup) bootstrapMetrics(c *gin.Context) {
	metrics, err := ng.getFacade().StatusMetrics().BootstrapMetrics()
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
			Data:  gin.H{"metrics": metrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// connectedPeersRatings returns the node's connected peers ratings
func (ng *nodeGroup) connectedPeersRatings(c *gin.Context) {
	ratings, err := ng.getFacade().GetConnectedPeersRatingsOnMainNetwork()
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
			Data:  gin.H{"ratings": ratings},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// managedKeysCount returns the node's number of managed keys
func (ng *nodeGroup) managedKeysCount(c *gin.Context) {
	count := ng.getFacade().GetManagedKeysCount()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"count": count},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// managedKeys returns all keys managed by the current node
func (ng *nodeGroup) managedKeys(c *gin.Context) {
	keys := ng.getFacade().GetManagedKeys()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"managedKeys": keys},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// managedKeysEligible returns the node's eligible managed keys
func (ng *nodeGroup) managedKeysEligible(c *gin.Context) {
	keys, err := ng.getFacade().GetEligibleManagedKeys()
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetEligibleManagedKeys, err)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"eligibleKeys": keys},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// managedKeysWaiting returns the node's waiting managed keys
func (ng *nodeGroup) managedKeysWaiting(c *gin.Context) {
	keys, err := ng.getFacade().GetWaitingManagedKeys()
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetWaitingManagedKeys, err)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"waitingKeys": keys},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
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
