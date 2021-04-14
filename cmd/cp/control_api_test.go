package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/naoki9911/CREBAS/pkg/capability"
)

var router = setupRouter()

func clearAll() {
	caps.Clear()
	capReqs.Clear()
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
	capReqRes := CapReqResponse{}
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
	capReqRes := CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)

	assert.Equal(t, capReq.RequestID, capReqRes.Request.RequestID)
	assert.Equal(t, len(capReqRes.GrantedCapabilities), 1)
	grantedCap := capReqRes.GrantedCapabilities[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)
}
