package main

import (
	"encoding/json"
	"fmt"
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
	r.POST("/capReq", postCapabilityRequest)

	return r
}

func postCapability(c *gin.Context) {
	var req []capability.Capability
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reqJSON, _ := json.Marshal(req)
	fmt.Println(string(reqJSON))

	for idx := range req {
		caps.Add(&req[idx])
	}

	c.JSON(http.StatusOK, req)
}

func postCapabilityRequest(c *gin.Context) {
	var req []capability.CapabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reqJSON, _ := json.Marshal(req)
	fmt.Println(string(reqJSON))

	for idx := range req {
		capReqs.Add(&req[idx])
	}

	c.JSON(http.StatusOK, req)
}
