package capability

import (
	"strings"

	"github.com/google/uuid"
)

// Capability is a capability
type Capability struct {
	CapabilityID          uuid.UUID                      `json:"capabilityID"`
	AssignerID            uuid.UUID                      `json:"assignerID"`
	AssigneeID            uuid.UUID                      `json:"assigneeID"`
	AppID                 uuid.UUID                      `json:"appID"`
	CapabilityName        string                         `json:"capabilityName"`
	CapabilityValue       string                         `json:"capabilityValue"`
	GrantCondition        string                         `json:"grantCondition,omitempty"`
	GrantPolicy           CapabilityAttributeBasedPolicy `json:"grantPolicy,omitempty"`
	AuthorizeCapabilityID uuid.UUID                      `json:"authorizeCapabilityID"`
	CapabilitySignature   CapabilitySignature            `json:"capabilitySignature"`
	GrantType             string                         `json:"grantType,omitempty"`
}

// CapabilityAttributeBasedPolicy is a condition for Capability
type CapabilityAttributeBasedPolicy struct {
	Condition              string    `json:"condition"`
	RequesterAttribute     string    `json:"requesterAttribute"`
	RequesterDeviceID      uuid.UUID `json:"requesterDeviceID,omitempty"`
	RequesterVendorID      uuid.UUID `json:"requesterVendorID,omitempty"`
	RequestCapabilityValue string    `json:"requestCapabilityValue,omitempty"`
}

// CapabilityRequest is a request for Capability
type CapabilityRequest struct {
	RequestID              uuid.UUID           `json:"requestID"`
	RequesterID            uuid.UUID           `json:"requesterID"`
	RequesteeID            uuid.UUID           `json:"requesteeID"`
	RequestCapabilityName  string              `json:"requestCapability"`
	RequestCapabilityValue string              `json:"requestCapabilityValue"`
	RequestSignature       CapabilitySignature `json:"requestSignature"`
	CapabilityID           uuid.UUID           `json:"capabilityID"`
}

// CapabilitySignature is a signature for capability
type CapabilitySignature struct {
	SignerID  uuid.UUID `json:"signerID"`
	SigneeID  uuid.UUID `json:"signeeID"`
	Signature string    `json:"signature"`
}

// UserGrantPolicy is a policy for user customized
type UserGrantPolicy struct {
	CapabilityID   uuid.UUID   `json:"capabilityID"`
	AutoGrant      bool        `json:"autoGrant"`
	AllowedUsersID []uuid.UUID `json:"allowedUsersID"`
}

const (
	CAPABILITY_NAME_EXTERNAL_COMMUNICATION = "ExternalCommunication"
)

func NewCreateSkeltonCapability() *Capability {
	cap := new(Capability)
	id, _ := uuid.NewRandom()
	cap.CapabilityID = id
	cap.AuthorizeCapabilityID = id
	id, _ = uuid.NewRandom()
	cap.AssignerID = id
	id, _ = uuid.NewRandom()
	cap.AssigneeID = id
	id, _ = uuid.NewRandom()
	cap.AppID = id
	cap.CapabilityName = "test-cap"
	cap.CapabilityValue = "test-cap-value"

	return cap
}

func NewCreateSkeltonCapabilityRequest() *CapabilityRequest {
	cap := new(CapabilityRequest)
	id, _ := uuid.NewRandom()
	cap.RequestID = id
	id, _ = uuid.NewRandom()
	cap.RequesterID = id
	id, _ = uuid.NewRandom()
	cap.RequesteeID = id
	id, _ = uuid.NewRandom()
	cap.CapabilityID = id
	cap.RequestCapabilityName = "test-cap"
	cap.RequestCapabilityValue = "test-cap-value"

	return cap
}

func (cap *Capability) IsDomainAllowed(domain string) bool {
	allowedDomain := cap.CapabilityValue

	if allowedDomain != "*" {
		// Matching For "*.example.com" or "*example.com"
		if strings.HasPrefix(allowedDomain, "*") {
			var matchDomain string
			if allowedDomain[1] == '.' {
				matchDomain = allowedDomain[2:]
			} else {
				matchDomain = allowedDomain[1:]
			}

			if !strings.HasSuffix(domain, matchDomain) {
				return false
			}
		}
	}

	return true

}

func (cap *Capability) GetGrantedCap(cpID uuid.UUID, capReq *CapabilityRequest) *Capability {
	if capReq.RequestCapabilityName == CAPABILITY_NAME_EXTERNAL_COMMUNICATION {
		return cap.getExternalCommunicationGrantedCap(cpID, capReq)
	}

	capID, _ := uuid.NewRandom()
	grantedCap := Capability{
		CapabilityID:          capID,
		AssignerID:            cpID,
		AssigneeID:            capReq.RequesterID,
		AppID:                 cap.AppID,
		CapabilityName:        capReq.RequestCapabilityName,
		CapabilityValue:       capReq.RequestCapabilityValue,
		AuthorizeCapabilityID: cap.CapabilityID,
		CapabilitySignature: CapabilitySignature{
			SignerID:  cpID,
			SigneeID:  capReq.RequesterID,
			Signature: "",
		},
		GrantCondition: "none",
	}

	return &grantedCap

}

func (cap *Capability) getExternalCommunicationGrantedCap(cpID uuid.UUID, capReq *CapabilityRequest) *Capability {
	if !cap.IsDomainAllowed(capReq.RequestCapabilityValue) {
		return nil
	}

	capID, _ := uuid.NewRandom()
	grantedCap := Capability{
		CapabilityID:          capID,
		AssignerID:            cpID,
		AssigneeID:            capReq.RequesterID,
		AppID:                 cap.AppID,
		CapabilityName:        capReq.RequestCapabilityName,
		CapabilityValue:       capReq.RequestCapabilityValue,
		AuthorizeCapabilityID: cap.CapabilityID,
		CapabilitySignature: CapabilitySignature{
			SignerID:  cpID,
			SigneeID:  capReq.RequesterID,
			Signature: "",
		},
		GrantCondition: "none",
	}

	return &grantedCap
}

func GetAutoGrantedCap(caps *CapabilityCollection, cpID uuid.UUID, capReq *CapabilityRequest) CapabilitySlice {
	candidateCaps := caps.Where(func(a *Capability) bool {
		return a.CapabilityName == capReq.RequestCapabilityName
	})

	grantedCaps := CapabilitySlice{}

	for idx := range candidateCaps {
		cap := candidateCaps[idx]
		if cap.GrantCondition == "always" {
			grantedCap := cap.GetGrantedCap(cpID, capReq)
			if grantedCap == nil {
				continue
			}
			grantedCaps = append(grantedCaps, grantedCap)
		}
	}

	return grantedCaps
}