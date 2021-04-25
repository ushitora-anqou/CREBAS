package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/capability"
)

var router = setupRouter()

func clearAll() {
	caps.Clear()
	capReqs.Clear()
	grantedCaps.Clear()
}

func TestPostCapability(t *testing.T) {
	clearAll()
	defer clearAll()

	cap1 := capability.NewCreateSkeltonCapability()
	cap2 := capability.NewCreateSkeltonCapability()
	capsReq := []*capability.Capability{cap1, cap2}
	reqBytes, err := json.Marshal(capsReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cap", bodyReader)
	router.ServeHTTP(w, req)

	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/cap", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capsRes := []capability.Capability{}
	json.Unmarshal(resbody, &capsRes)

	assert.Equal(t, len(capsRes), 2)
	assert.Equal(t, capsRes[0].CapabilityID, cap1.CapabilityID)
	assert.Equal(t, capsRes[1].CapabilityID, cap2.CapabilityID)

	assert.Equal(t, len(caps.GetAll()), 2)
	cap1Test := caps.GetByIndex(0)
	cap2Test := caps.GetByIndex(1)
	assert.Equal(t, cap1Test.CapabilityID, cap1.CapabilityID)
	assert.Equal(t, cap2Test.CapabilityID, cap2.CapabilityID)
}

func TestPostCapabilityReqest(t *testing.T) {
	clearAll()
	defer clearAll()

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	reqBytes, err := json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capReqRes := capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
}

func TestAutoGrant(t *testing.T) {
	clearAll()
	defer clearAll()

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"

	capsReq := []*capability.Capability{cap1, cap2}
	reqBytes, err := json.Marshal(capsReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cap", bodyReader)
	router.ServeHTTP(w, req)

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	reqBytes, err = json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capReqRes := capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 1)
	grantedCap := capReqRes.GrantedCapabilities[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)
}

func TestManualGrant(t *testing.T) {
	clearAll()
	defer clearAll()

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"

	capsReq := []*capability.Capability{cap1, cap2}
	reqBytes, err := json.Marshal(capsReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cap", bodyReader)
	router.ServeHTTP(w, req)

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	reqBytes, err = json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capReqRes := capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 1)
	grantedCap := capReqRes.GrantedCapabilities[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)

	bodyReader = strings.NewReader("")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq/"+capReq.RequestID.String()+"/grant/"+cap2.CapabilityID.String(), bodyReader)
	router.ServeHTTP(w, req)

	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	capReqRes = capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	for _, grantCap := range capReqRes.GrantedCapabilities {
		fmt.Println(grantCap)
	}
	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 2)
	testCaps := capReqRes.GrantedCapabilities.Where(func(c *capability.Capability) bool {
		return c.AuthorizeCapabilityID == cap2.CapabilityID
	})
	assert.Equal(t, len(testCaps), 1)
	grantedCap = testCaps[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap2.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)
}

func TestUserGrant(t *testing.T) {
	clearAll()
	defer clearAll()

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"

	capsReq := []*capability.Capability{cap1, cap2}
	reqBytes, err := json.Marshal(capsReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cap", bodyReader)
	router.ServeHTTP(w, req)

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	reqBytes, err = json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capReqRes := capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 1)
	grantedCap := capReqRes.GrantedCapabilities[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)

	policyID, _ := uuid.NewRandom()
	userPolicy := capability.UserGrantPolicy{
		UserGrantPolicyID: policyID,
		CapabilityID:      cap2.CapabilityID,
		Grant:             true,
		RequesterID:       capReq.RequesterID,
	}
	reqBytes, err = json.Marshal(userPolicy)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/user/grantPolicy", bodyReader)
	router.ServeHTTP(w, req)

	reqBytes, err = json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	capReqRes = capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	for _, grantCap := range capReqRes.GrantedCapabilities {
		fmt.Println(grantCap)
	}
	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 2)
	testCaps := capReqRes.GrantedCapabilities.Where(func(c *capability.Capability) bool {
		return c.AuthorizeCapabilityID == cap2.CapabilityID
	})
	assert.Equal(t, len(testCaps), 1)
	grantedCap = testCaps[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap2.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)
}

func TestUserGrant2(t *testing.T) {
	clearAll()
	defer clearAll()

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"

	capsReq := []*capability.Capability{cap1, cap2}
	reqBytes, err := json.Marshal(capsReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cap", bodyReader)
	router.ServeHTTP(w, req)

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	reqBytes, err = json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capReqRes := capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 1)
	grantedCap := capReqRes.GrantedCapabilities[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)

	policyID, _ := uuid.NewRandom()
	userPolicy := capability.UserGrantPolicy{
		UserGrantPolicyID: policyID,
		CapabilityID:      cap2.CapabilityID,
		Grant:             false,
		RequesterID:       capReq.RequesterID,
	}
	reqBytes, err = json.Marshal(userPolicy)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/user/grantPolicy", bodyReader)
	router.ServeHTTP(w, req)

	reqBytes, err = json.Marshal(capReq)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	capReqRes = capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	for _, grantCap := range capReqRes.GrantedCapabilities {
		fmt.Println(grantCap)
	}
	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 1)
}
