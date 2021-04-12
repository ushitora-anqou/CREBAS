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

	capReq1 := capability.NewCreateSkeltonCapabilityRequest()
	capReq2 := capability.NewCreateSkeltonCapabilityRequest()
	capReqsReq := []*capability.CapabilityRequest{capReq1, capReq2}
	reqBytes, err := json.Marshal(capReqsReq)
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
	capReqsRes := []capability.CapabilityRequest{}
	json.Unmarshal(resbody, &capReqsRes)

	assert.Equal(t, len(capReqsRes), 2)
	assert.Equal(t, capReqsRes[0].CapabilityID, capReq1.CapabilityID)
	assert.Equal(t, capReqsRes[1].CapabilityID, capReq2.CapabilityID)

	assert.Equal(t, len(capReqs.GetAll()), 2)
	capReq1Test := capReqs.GetByIndex(0)
	capReq2Test := capReqs.GetByIndex(1)
	assert.Equal(t, capReq1Test.CapabilityID, capReq1.CapabilityID)
	assert.Equal(t, capReq2Test.CapabilityID, capReq2.CapabilityID)
}
