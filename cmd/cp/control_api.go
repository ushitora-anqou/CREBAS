package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/naoki9911/CREBAS/pkg/capability"
)

func StartAPIServer() error {
	return setupRouter().Run("0.0.0.0:8081")
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/cap", postCapability)
	r.GET("/cap", getCapability)
	r.POST("/capReq", postCapabilityRequest)
	r.GET("/capReq", getCapabilityRequest)

	return r
}

func postCapability(c *gin.Context) {
	var req []capability.Capability
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for idx := range req {
		caps.Add(&req[idx])
	}

	c.JSON(http.StatusOK, req)
}

func getCapability(c *gin.Context) {
	c.JSON(http.StatusOK, caps.GetAll())
}

type CapReqResponse struct {
	Request             capability.CapabilityRequest `json:"request"`
	GrantedCapabilities capability.CapabilitySlice   `json:"grantedCapabilities"`
}

func postCapabilityRequest(c *gin.Context) {
	var req capability.CapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grantCaps := capability.GetAutoGrantedCap(caps, config.cpID, &req)

	res := CapReqResponse{
		Request:             req,
		GrantedCapabilities: grantCaps,
	}

	c.JSON(http.StatusOK, res)

	for idx := range grantCaps {
		grantedCaps.Add(grantCaps[idx])
	}
}

func getCapabilityRequest(c *gin.Context) {
	c.JSON(http.StatusOK, grantedCaps.GetAll())
}
