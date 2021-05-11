package main

import (
	"crypto/rsa"
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
var appCerts = capability.NewAppCertificateCollection()

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
	r.POST("/app/cert", postAppCert)
	r.GET("/app/cpCert", getCPCert)
	r.GET("/app/userCert", getUserCert)
	r.POST("/cap", postCapability)
	r.GET("/cap", getCapability)
	r.GET("/cap/granted", getGrantedCapability)
	r.GET("/cap/delegated", getDelegatedCapability)
	r.POST("/capReq", postCapabilityRequest)
	r.GET("/capReq", getCapabilityRequest)
	r.GET("/capReq/pending", getPendingCapabilityRequest)
	r.POST("/capReq/:reqID/grant/:capID", postCapabilityRequestGrantManually)
	r.POST("/user/grantPolicy", postUserGrantPolicy)

	return r
}

func postAppCert(c *gin.Context) {
	var req capability.AppCertificate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := req.Decode()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = capability.VerifyCertificate(req.Certificate, config.caCert)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appCerts.Add(&req)
	c.JSON(http.StatusOK, req)
}

func getCPCert(c *gin.Context) {
	c.JSON(http.StatusOK, config.cpCert)
}

func getUserCert(c *gin.Context) {
	c.JSON(http.StatusOK, config.userCert)
}

func postCapability(c *gin.Context) {
	var req []capability.Capability
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for idx := range req {
		cap := req[idx]
		appCert := appCerts.GetByID(cap.AssignerID)
		if appCert == nil {
			c.JSON(http.StatusBadRequest, "appCert "+cap.AssignerID.String()+"not found")
			return
		}
		err := cap.Verify(appCert.Certificate.PublicKey.(*rsa.PublicKey))
		if err != nil {
			c.JSON(http.StatusBadRequest, "verify failed")
			return
		}
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

func getGrantedCapability(c *gin.Context) {
	c.JSON(http.StatusOK, grantedCaps.GetAll())
}

func getDelegatedCapability(c *gin.Context) {
	delegatedCaps := caps.Where(func(c *capability.Capability) bool {
		return c.CapabilityID == c.AuthorizeCapabilityID
	})

	c.JSON(http.StatusOK, delegatedCaps)
}

func postCapabilityRequest(c *gin.Context) {
	req := &capability.CapabilityRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appCert := appCerts.GetByID(req.RequesterID)
	if appCert == nil {
		c.JSON(http.StatusBadRequest, "appCert "+req.RequesterID.String()+"not found")
		return
	}

	err := req.Verify(appCert.Certificate.PublicKey.(*rsa.PublicKey))
	if err != nil {
		c.JSON(http.StatusBadRequest, "verify failed")
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

func getPendingCapabilityRequest(c *gin.Context) {
	capReqAll := capReqs.GetAll()
	delegatedCaps := caps.Where(func(c *capability.Capability) bool {
		return c.CapabilityID == c.AuthorizeCapabilityID
	})
	pendingCapReqs := []capability.CapReqPendingResponse{}
	for idx := range capReqAll {
		pendingCapReq := capability.CapReqPendingResponse{}
		capReq := capReqAll[idx]
		pendingCapReq.Request = *capReq
		candidateCaps := delegatedCaps.Where(func(c *capability.Capability) bool {
			return c.CapabilityName == capReq.RequestCapabilityName
		})

		for idx := range candidateCaps {
			alreadyGrantedCaps := grantedCaps.Where(func(c *capability.Capability) bool {
				return c.AppID == candidateCaps[idx].AppID
			})
			if len(alreadyGrantedCaps) != 0 {
				continue
			}

			pendingCapReq.PendingCapabilities = append(pendingCapReq.PendingCapabilities, candidateCaps[idx])
		}
		pendingCapReqs = append(pendingCapReqs, pendingCapReq)
	}
	c.JSON(http.StatusOK, pendingCapReqs)
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

	capDelegatedToUser := cap.GetDelegatedCapability(config.cpID, config.userID)
	grantedCaps.Add(capDelegatedToUser)
	grantCap := capDelegatedToUser.GetGrantedCap(config.userID, capReq)

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
