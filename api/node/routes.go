package node

import (
	"net/http"
	"net/url"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok", "running": ef.IsNodeRunning()})
}

// StartNode will start the node instance
func StartNode(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	if ef.IsNodeRunning() {
		c.JSON(http.StatusOK, gin.H{"message": "Node already running"})
		return
	}

	err := ef.StartNode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bad init of node: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// Address returns the information about the address passed as parameter
func Address(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	currentAddress := ef.GetCurrentPublicKey()
	address, err := url.Parse(currentAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cound not parse node's public key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"address": address.String()})
}

// StopNode will stop the node instance
func StopNode(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	if !ef.IsNodeRunning() {
		c.JSON(http.StatusOK, gin.H{"message": "Node already stopped"})
		return
	}

	err := ef.StopNode()
	if err != nil && ef.IsNodeRunning() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not stop node: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
