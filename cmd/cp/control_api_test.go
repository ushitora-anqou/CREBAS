package main

import (
	"crypto/rsa"
	"encoding/base64"
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
	appCerts.Clear()
}

func TestPostCapability(t *testing.T) {
	clearAll()
	defer clearAll()

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap1 := capability.NewCreateSkeltonCapability()
	cap2 := capability.NewCreateSkeltonCapability()
	cap1.AssignerID = assignerID
	cap2.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap2.AssigneeID = config.cpID
	cap1.Sign(privKey)
	cap2.Sign(privKey)
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

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.Sign(privKey)
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

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.Sign(privKey)

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
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)
}

func TestAutoGrant2(t *testing.T) {
	clearAll()
	defer clearAll()

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	deviceID, _ := uuid.NewRandom()
	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "conditional"
	cap1.GrantPolicy = capability.CapabilityAttributeBasedPolicy{
		RequesterAttribute: "DeviceID",
		RequesterDeviceID:  deviceID,
	}
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.Sign(privKey)

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
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.DeviceID = deviceID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)
}

func TestAutoGrant3(t *testing.T) {
	clearAll()
	defer clearAll()

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	vendorID, _ := uuid.NewRandom()
	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "conditional"
	cap1.GrantPolicy = capability.CapabilityAttributeBasedPolicy{
		RequesterAttribute: "VendorID",
		RequesterVendorID:  vendorID,
	}
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.Sign(privKey)

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
	capReq.VendorID = vendorID
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
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

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.Sign(privKey)

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
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
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
	assert.Equal(t, len(testCaps), 0)
	testCaps = capReqRes.GrantedCapabilities.Where(func(c *capability.Capability) bool {
		return c.AssignerID == config.userID
	})
	assert.Equal(t, len(testCaps), 1)
	grantedCap = testCaps[0]
	assert.Equal(t, nil, grantedCap.Verify(config.userCert.Certificate.PublicKey.(*rsa.PublicKey)))
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AssignerID, config.userID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)
}

func TestUserGrant(t *testing.T) {
	clearAll()
	defer clearAll()

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.Sign(privKey)

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
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap2.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)
}

func TestUserGrant2(t *testing.T) {
	clearAll()
	defer clearAll()

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.Sign(privKey)

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
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
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

func TestPendingCapReqAndDelegatedCap(t *testing.T) {
	clearAll()
	defer clearAll()

	assignerID, _ := uuid.NewRandom()
	err := postTestCert(assignerID)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	privKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"
	cap1.AssignerID = assignerID
	cap1.AssigneeID = config.cpID
	cap1.Sign(privKey)

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"
	cap2.AssignerID = assignerID
	cap2.AssigneeID = config.cpID
	cap2.Sign(privKey)

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
	capReq.RequesterID = assignerID
	capReq.RequesteeID = config.cpID
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"
	capReq.Sign(privKey)
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
	assert.Equal(t, nil, grantedCap.Verify(config.cpCert.Certificate.PublicKey.(*rsa.PublicKey)))
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AuthorizeCapabilityID, cap1.CapabilityID)
	assert.Equal(t, grantedCap.AssignerID, config.cpID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	assert.Equal(t, grantedCaps.Count(), 1)
	grantedCapCP := grantedCaps.GetByIndex(0)
	assert.Equal(t, grantedCapCP.CapabilityID, grantedCap.CapabilityID)

	req = httptest.NewRequest("GET", "/capReq/pending", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	pendingCapReqs := []capability.CapReqPendingResponse{}
	json.Unmarshal(resbody, &pendingCapReqs)

	assert.Equal(t, len(pendingCapReqs), 1)
	assert.Equal(t, pendingCapReqs[0].Request.RequestID, capReq.RequestID)
	assert.Equal(t, len(pendingCapReqs[0].PendingCapabilities), 1)
	assert.Equal(t, pendingCapReqs[0].PendingCapabilities[0].CapabilityID, cap2.CapabilityID)

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
	assert.Equal(t, len(testCaps), 0)
	testCaps = capReqRes.GrantedCapabilities.Where(func(c *capability.Capability) bool {
		return c.AssignerID == config.userID
	})
	assert.Equal(t, len(testCaps), 1)
	grantedCap = testCaps[0]
	assert.Equal(t, grantedCap.CapabilityName, capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION)
	assert.Equal(t, grantedCap.CapabilityValue, capReq.RequestCapabilityValue)
	assert.Equal(t, grantedCap.AssignerID, config.userID)
	assert.Equal(t, grantedCap.AssigneeID, capReq.RequesterID)

	req = httptest.NewRequest("GET", "/capReq/pending", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	pendingCapReqs = []capability.CapReqPendingResponse{}
	json.Unmarshal(resbody, &pendingCapReqs)

	assert.Equal(t, len(pendingCapReqs), 1)
	assert.Equal(t, pendingCapReqs[0].Request.RequestID, capReq.RequestID)
	assert.Equal(t, len(pendingCapReqs[0].PendingCapabilities), 0)

	req = httptest.NewRequest("GET", "/cap/delegated", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	delegatedCaps := []capability.Capability{}
	json.Unmarshal(resbody, &delegatedCaps)

	assert.Equal(t, len(delegatedCaps), 2)
	if delegatedCaps[0].CapabilityID == cap1.CapabilityID {
		assert.Equal(t, delegatedCaps[0].CapabilityID, cap1.CapabilityID)
		assert.Equal(t, delegatedCaps[1].CapabilityID, cap2.CapabilityID)
	} else {
		assert.Equal(t, delegatedCaps[0].CapabilityID, cap2.CapabilityID)
		assert.Equal(t, delegatedCaps[1].CapabilityID, cap1.CapabilityID)
	}
}

func TestPostCertificate(t *testing.T) {
	appCerts.Clear()
	defer appCerts.Clear()

	id, _ := uuid.NewRandom()
	certBytes, err := capability.ReadCertificateWithoutDecode("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	certBase64 := base64.StdEncoding.EncodeToString(certBytes)
	appCert := capability.AppCertificate{
		AppID:             id,
		CertificateString: certBase64,
	}

	reqBytes, err := json.Marshal(appCert)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/app/cert", bodyReader)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	appCertTest := appCerts.GetByIndex(0)
	err = capability.VerifyCertificate(appCertTest.Certificate, config.caCert)
	assert.Equal(t, err, nil)
}

func postTestCert(appID uuid.UUID) error {
	certBytes, err := capability.ReadCertificateWithoutDecode("/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.crt")
	if err != nil {
		return err
	}

	certBase64 := base64.StdEncoding.EncodeToString(certBytes)
	appCert := capability.AppCertificate{
		AppID:             appID,
		CertificateString: certBase64,
	}

	reqBytes, err := json.Marshal(appCert)
	if err != nil {
		return err
	}
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/app/cert", bodyReader)
	router.ServeHTTP(w, req)

	return nil
}

func TestGetCPCert(t *testing.T) {
	req := httptest.NewRequest("GET", "/app/cpCert", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	appCert := capability.AppCertificate{}
	json.Unmarshal(resbody, &appCert)

	err := appCert.Decode()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = capability.VerifyCertificate(appCert.Certificate, config.caCert)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}

func TestGetUserCert(t *testing.T) {
	req := httptest.NewRequest("GET", "/app/userCert", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	appCert := capability.AppCertificate{}
	json.Unmarshal(resbody, &appCert)

	err := appCert.Decode()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = capability.VerifyCertificate(appCert.Certificate, config.caCert)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}
