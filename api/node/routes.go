package node

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	IsNodeRunning() bool
	StartNode() error
	StopNode() error
}

// Routes defines node related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/start", StartNode)
	router.GET("/status", Status)
	router.GET("/stop", StopNode)
}

// Status returns the state of the node e.g. running/stopped
func Status(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid app context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok", "running": ef.IsNodeRunning()})
}

// StartNode will start the node instance
func StartNode(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid app context"})
		return
	}

	if ef.IsNodeRunning() {
		c.JSON(http.StatusOK, gin.H{"message": "Node already running"})
		return
	}

	err := ef.StartNode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Bad init of node: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// StopNode will stop the node instance
func StopNode(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid app context"})
		return
	}

	if !ef.IsNodeRunning() {
		c.JSON(http.StatusOK, gin.H{"message": "Node already stopped"})
		return
	}

	err := ef.StopNode()
	if err != nil && ef.IsNodeRunning() {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not stop node: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}