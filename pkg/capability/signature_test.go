package capability

import (
	"crypto/rsa"
	"encoding/base64"
	"testing"
)

func TestReadPrivateKey(t *testing.T) {
	_, err := ReadPrivateKey("/home/naoki/CREBAS/test/keys/cp/test-cp.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}

func TestReadCertificate(t *testing.T) {
	_, err := ReadCertificate("/home/naoki/CREBAS/test/keys/ca/test-ca.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}

func TestVerifyCertificate(t *testing.T) {
	cacert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/ca/test-ca.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = VerifyCertificate(cert, cacert)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}

func TestCapabilitySignAndVerify(t *testing.T) {
	privKey, err := ReadPrivateKey("/home/naoki/CREBAS/test/keys/cp/test-cp.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cert2, err := ReadCertificate("/home/naoki/CREBAS/test/keys/user/test-user.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cap := NewCreateSkeltonCapability()
	err = cap.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = cap.Verify(cert.PublicKey.(*rsa.PublicKey))
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = cap.Verify(cert2.PublicKey.(*rsa.PublicKey))
	if err == nil {
		t.Fatalf("Failed")
	}
}

func TestCapabilitySignAndVerify2(t *testing.T) {
	privKey, err := ReadPrivateKey("/home/naoki/CREBAS/test/keys/cp/test-cp.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cap := NewCreateSkeltonCapability()
	err = cap.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cap.CapabilityName = "invalid"
	err = cap.Verify(cert.PublicKey.(*rsa.PublicKey))
	if err == nil {
		t.Fatalf("Failed")
	}
}

func TestCapabilityRequestSignAndVerify(t *testing.T) {
	privKey, err := ReadPrivateKey("/home/naoki/CREBAS/test/keys/cp/test-cp.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cert2, err := ReadCertificate("/home/naoki/CREBAS/test/keys/user/test-user.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cap := NewCreateSkeltonCapabilityRequest()
	err = cap.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = cap.Verify(cert.PublicKey.(*rsa.PublicKey))
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = cap.Verify(cert2.PublicKey.(*rsa.PublicKey))
	if err == nil {
		t.Fatalf("Failed")
	}
}

func TestCapabilityRequestSignAndVerify2(t *testing.T) {
	privKey, err := ReadPrivateKey("/home/naoki/CREBAS/test/keys/cp/test-cp.key")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cap := NewCreateSkeltonCapabilityRequest()
	err = cap.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	cap.RequestCapabilityName = "invalid"

	err = cap.Verify(cert.PublicKey.(*rsa.PublicKey))
	if err == nil {
		t.Fatalf("Failed")
	}
}

func TestBase64Certificate(t *testing.T) {
	caCert, err := ReadCertificate("/home/naoki/CREBAS/test/keys/ca/test-ca.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	certBytes, err := ReadCertificateWithoutDecode("/home/naoki/CREBAS/test/keys/cp/test-cp.crt")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	certBase64 := base64.StdEncoding.EncodeToString(certBytes)

	certString, err := base64.StdEncoding.DecodeString(certBase64)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cert, err := DecodeCertificate([]byte(certString))
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = VerifyCertificate(cert, caCert)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}
