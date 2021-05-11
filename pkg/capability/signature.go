package capability

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
)

type AppCertificate struct {
	AppID             uuid.UUID         `json:"appID"`
	CertificateString string            `json:"certificate"`
	Certificate       *x509.Certificate `json:"-"`
}

func (c *AppCertificate) Decode() error {
	certString, err := base64.StdEncoding.DecodeString(c.CertificateString)
	if err != nil {
		return err
	}

	cert, err := DecodeCertificate([]byte(certString))
	if err != nil {
		return nil
	}

	c.Certificate = cert

	return nil
}

// Sign signs capability
func (cap *Capability) Sign(privateKey *rsa.PrivateKey) error {

	signCap := Capability{
		AssignerID:      cap.AssigneeID,
		AssigneeID:      cap.AssigneeID,
		CapabilityID:    cap.CapabilityID,
		AppID:           cap.AppID,
		CapabilityName:  cap.CapabilityName,
		CapabilityValue: cap.CapabilityValue,
		GrantCondition:  cap.GrantCondition,
		GrantPolicy:     cap.GrantPolicy,
		CapabilitySignature: CapabilitySignature{
			SignerID:  cap.CapabilitySignature.SignerID,
			SigneeID:  cap.CapabilitySignature.SigneeID,
			Signature: "",
		},
	}

	h := crypto.Hash.New(crypto.SHA256)
	h.Write(([]byte)(fmt.Sprintf("%v", signCap)))
	hash := h.Sum(nil)

	signedData, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash)
	if err != nil {
		return err
	}

	signature := base64.StdEncoding.EncodeToString(signedData)
	cap.CapabilitySignature.Signature = signature
	return nil
}

// Verify verifies capability
func (cap *Capability) Verify(publicKey *rsa.PublicKey) error {
	signDataByte, err := base64.StdEncoding.DecodeString(cap.CapabilitySignature.Signature)
	if err != nil {
		return err
	}

	signCap := Capability{
		AssignerID:      cap.AssigneeID,
		AssigneeID:      cap.AssigneeID,
		CapabilityID:    cap.CapabilityID,
		AppID:           cap.AppID,
		CapabilityName:  cap.CapabilityName,
		CapabilityValue: cap.CapabilityValue,
		GrantCondition:  cap.GrantCondition,
		GrantPolicy:     cap.GrantPolicy,
		CapabilitySignature: CapabilitySignature{
			SignerID:  cap.CapabilitySignature.SignerID,
			SigneeID:  cap.CapabilitySignature.SigneeID,
			Signature: "",
		},
	}

	h := crypto.Hash.New(crypto.SHA256)
	h.Write(([]byte)(fmt.Sprintf("%v", signCap)))
	hash := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash, signDataByte)
	if err != nil {
		return err
	}
	return nil
}

// Sign signs capability request
func (capReq *CapabilityRequest) Sign(privateKey *rsa.PrivateKey) error {
	capReq.RequestSignature.SignerID = capReq.RequesterID
	capReq.RequestSignature.SigneeID = capReq.RequesteeID
	capReq.RequestSignature.Signature = ""

	h := crypto.Hash.New(crypto.SHA256)
	h.Write(([]byte)(fmt.Sprintf("%v", capReq)))
	hash := h.Sum(nil)

	signedData, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash)
	if err != nil {
		return err
	}

	signature := base64.StdEncoding.EncodeToString(signedData)
	capReq.RequestSignature.Signature = signature
	return nil
}

// Verify verifies capability request
func (capReq *CapabilityRequest) Verify(publicKey *rsa.PublicKey) error {
	reqSig := capReq.RequestSignature.Signature
	signDataByte, err := base64.StdEncoding.DecodeString(reqSig)
	if err != nil {
		return err
	}
	capReq.RequestSignature.Signature = ""

	h := crypto.Hash.New(crypto.SHA256)
	h.Write(([]byte)(fmt.Sprintf("%v", capReq)))
	hash := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash, signDataByte)
	capReq.RequestSignature.Signature = reqSig
	if err != nil {
		return err
	}

	return nil
}

// ReadPrivateKey read privateKey from file
func ReadPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid private key data")
	}

	var key *rsa.PrivateKey
	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else if block.Type == "PRIVATE KEY" {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not RSA private key")
		}
	} else {
		return nil, fmt.Errorf("invalid private key type : %s", block.Type)
	}

	key.Precompute()

	if err := key.Validate(); err != nil {
		return nil, err
	}

	return key, nil
}

// ReadCertificate read certificate from file
func ReadCertificate(certPath string) (*x509.Certificate, error) {
	bytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid certificate data")
	}

	return x509.ParseCertificate(block.Bytes)
}

func ReadCertificateWithoutDecode(certPath string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(certPath)

	return bytes, err
}

func DecodeCertificate(bytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid certificate data")
	}

	return x509.ParseCertificate(block.Bytes)
}

//VerifyCertificate verifies certificate with ca certificate
func VerifyCertificate(cert *x509.Certificate, caCert *x509.Certificate) error {
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	_, err := cert.Verify(opts)
	return err
}
