package node

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	IsNodeRunning() bool
	StartNode() error
	StopNode() error
	GetCurrentPublicKey() string
}

// Routes defines node related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/start", StartNode)
	router.GET("/status", Status)
	router.GET("/stop", StopNode)
	router.GET("/address", Address)
}

// Status returns the state of the node e.g. running/stopped
func Status(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok", "running": ef.IsNodeRunning()})
}

// StartNode will start the node instance
func StartNode(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	if ef.IsNodeRunning() {
		c.JSON(http.StatusOK, gin.H{"message": errors.ErrNodeAlreadyRunning.Error()})
		return
	}

	err := ef.StartNode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrBadInitOfNode.Error(), err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// Address returns the information about the address passed as parameter
func Address(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	currentAddress := ef.GetCurrentPublicKey()
	address, err := url.Parse(currentAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrCouldNotParsePubKey.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"address": address.String()})
}

// StopNode will stop the node instance
func StopNode(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	if !ef.IsNodeRunning() {
		c.JSON(http.StatusOK, gin.H{"message": errors.ErrNodeAlreadyStopped.Error()})
		return
	}

	err := ef.StopNode()
	if err != nil && ef.IsNodeRunning() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrCouldNotStopNode.Error(), err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
