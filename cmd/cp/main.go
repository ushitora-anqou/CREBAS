package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/capability"
)

type CPConfig struct {
	cpID uuid.UUID
}

func loadCPConfig() CPConfig {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	cpConfig := CPConfig{
		cpID: id,
	}

	return cpConfig
}

var config CPConfig = loadCPConfig()

func main() {
	log.Printf("info: Starting CapabilityProvider(cpID: %v)", config.cpID)

	router := setupRouter()
	addTestCaps(router)
	router.Run("0.0.0.0:8081")
	//StartAPIServer()
}

func addTestCaps(r *gin.Engine) {

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap2.CapabilityValue = "*.example.com"
	cap2.GrantCondition = "none"

	capsReq := []*capability.Capability{cap1, cap2}
	reqBytes, _ := json.Marshal(capsReq)
	bodyReader := strings.NewReader(string(reqBytes))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cap", bodyReader)
	r.ServeHTTP(w, req)

	capReq := capability.NewCreateSkeltonCapabilityRequest()
	capReq.RequestCapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	capReq.RequestCapabilityValue = "*.test.hoge.example.com"

	reqBytes, _ = json.Marshal(capReq)
	bodyReader = strings.NewReader(string(reqBytes))
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/capReq", bodyReader)
	r.ServeHTTP(w, req)

	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	capReqRes := capability.CapReqResponse{}
	json.Unmarshal(resbody, &capReqRes)
}
