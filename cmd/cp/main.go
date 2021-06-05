package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
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
	cpID        uuid.UUID
	userID      uuid.UUID
	caCert      *x509.Certificate
	cpPrivKey   *rsa.PrivateKey
	userPrivKey *rsa.PrivateKey
	cpCert      capability.AppCertificate
	userCert    capability.AppCertificate
}

func loadCPConfig() CPConfig {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	userId, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	caCert, err := capability.ReadCertificate("/home/naoki/CREBAS/test/keys/ca/test-ca.crt")
	if err != nil {
		panic(err)
	}

	cpPrivKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/cp/test-cp.key")
	if err != nil {
		panic(err)
	}

	userPrivKey, err := capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/user/test-user.key")
	if err != nil {
		panic(err)
	}

	cpCertBytes, err := capability.ReadCertificateWithoutDecode("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		panic(err)
	}
	cpCertBase64 := base64.StdEncoding.EncodeToString(cpCertBytes)
	cpCert := capability.AppCertificate{
		AppID:             id,
		CertificateString: cpCertBase64,
	}

	userCertBytes, err := capability.ReadCertificateWithoutDecode("/home/naoki/CREBAS/test/keys/user/test-user.crt")
	if err != nil {
		panic(err)
	}
	userCertBase64 := base64.StdEncoding.EncodeToString(userCertBytes)
	userCert := capability.AppCertificate{
		AppID:             userId,
		CertificateString: userCertBase64,
	}

	cpConfig := CPConfig{
		cpID:        id,
		userID:      userId,
		caCert:      caCert,
		cpPrivKey:   cpPrivKey,
		userPrivKey: userPrivKey,
		cpCert:      cpCert,
		userCert:    userCert,
	}

	return cpConfig
}

var config CPConfig = loadCPConfig()

func main() {
	log.Printf("info: Starting CapabilityProvider(cpID: %v)", config.cpID)

	router := setupRouter()
	//addTestCaps(router)
	router.Run("0.0.0.0:8081")
	//StartAPIServer()
}

func addTestCaps(r *gin.Engine) {

	cap1 := capability.NewCreateSkeltonCapability()
	cap1.AssigneeID = config.cpID
	cap1.AssignerID = config.cpID
	cap1.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap1.CapabilityValue = "*.hoge.example.com"
	cap1.GrantCondition = "always"

	cap2 := capability.NewCreateSkeltonCapability()
	cap2.AssigneeID = config.cpID
	cap2.AssignerID = config.cpID
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
