package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/capability"
)

var caps = capability.NewCapabilityCollection()
var grantedCaps = capability.NewCapabilityCollection()
var capReqs = capability.NewCapabilityRequestCollection()
var userGrantPolicies = capability.NewUserGrantPolicyCollection()

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
	r.POST("/capReq/:reqID/grant/:capID", postCapabilityRequestGrantManually)
	r.POST("/user/grantPolicy", postUserGrantPolicy)

	return r
}

func postCapability(c *gin.Context) {
	var req []capability.Capability
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for idx := range req {
		cap := req[idx]
		if !caps.Contains(&cap) {
			caps.Add(&cap)
		}
	}

	c.JSON(http.StatusOK, req)
}

func postUserGrantPolicy(c *gin.Context) {
	var req capability.UserGrantPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !userGrantPolicies.Contains(&req) {
		userGrantPolicies.Add(&req)
	}

	c.JSON(http.StatusOK, req)
}

func getCapability(c *gin.Context) {
	c.JSON(http.StatusOK, caps.GetAll())
}

func postCapabilityRequest(c *gin.Context) {
	req := &capability.CapabilityRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if capReqs.Contains(req) {
		req = capReqs.GetByID(req.RequestID)
	} else {
		capReqs.Add(req)
	}

	grantCaps := capability.GetUserAndManualGrantedCap(caps, config.cpID, req, userGrantPolicies)

	for idx := range grantCaps {
		grantCap := grantCaps[idx]
		alreadyGrantedCaps := grantedCaps.Where(func(c1 *capability.Capability) bool {
			return c1.AuthorizeCapabilityID == grantCap.AuthorizeCapabilityID && c1.CapabilityValue == grantCap.CapabilityValue
		})
		if len(alreadyGrantedCaps) != 0 {
			continue
		} else {
			grantedCaps.Add(grantCap)
		}

		alreadyGrantedCaps = req.GrantedCapabilities.Where(func(c1 *capability.Capability) bool {
			return c1.AuthorizeCapabilityID == grantCap.AuthorizeCapabilityID && c1.CapabilityValue == grantCap.CapabilityValue
		})
		if len(alreadyGrantedCaps) != 0 {
			continue
		} else {
			req.GrantedCapabilities.Add(grantCap)
		}
	}

	res := capability.CapReqResponse{
		Request:             *req,
		GrantedCapabilities: req.GrantedCapabilities.GetAll(),
	}

	c.JSON(http.StatusOK, res)

}

func getCapabilityRequest(c *gin.Context) {
	c.JSON(http.StatusOK, capReqs.GetAll())
}

func postCapabilityRequestGrantManually(c *gin.Context) {
	reqID, err := uuid.Parse(c.Param("reqID"))
	if err != nil {
		log.Printf("error: invalid id %v", reqID)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	capID, err := uuid.Parse(c.Param("capID"))
	if err != nil {
		log.Printf("error: invalid id %v", capID)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Printf("info: Grant Manually CapReqID: %v CapID: %v", reqID, capID)

	capReq := capReqs.GetByID(reqID)
	if capReq == nil {
		log.Printf("error: not found Capability Request %v", reqID)
		c.JSON(http.StatusBadRequest, "not found Capability Request "+reqID.String())
		return
	}
	cap := caps.GetByID(capID)
	if cap == nil {
		log.Printf("error: not found Capability %v", capID)
		c.JSON(http.StatusBadRequest, "not found Capability "+capID.String())
		return
	}

	grantCap := cap.GetGrantedCap(config.cpID, capReq)

	alreadyGrantedCaps := grantedCaps.Where(func(c1 *capability.Capability) bool {
		return c1.AuthorizeCapabilityID == grantCap.AuthorizeCapabilityID && c1.CapabilityValue == grantCap.CapabilityValue
	})
	if len(alreadyGrantedCaps) == 0 {
		grantedCaps.Add(grantCap)
	}

	alreadyGrantedCaps = capReq.GrantedCapabilities.Where(func(c1 *capability.Capability) bool {
		return c1.AuthorizeCapabilityID == grantCap.AuthorizeCapabilityID && c1.CapabilityValue == grantCap.CapabilityValue
	})
	if len(alreadyGrantedCaps) == 0 {
		capReq.GrantedCapabilities.Add(grantCap)
	}

	res := capability.CapReqResponse{
		Request:             *capReq,
		GrantedCapabilities: []*capability.Capability{grantCap},
	}

	c.JSON(http.StatusOK, res)

}
