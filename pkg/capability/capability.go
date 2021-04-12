package capability

import "github.com/google/uuid"

// Capability is a capability
type Capability struct {
	CapabilityID          uuid.UUID                      `json:"capabilityID"`
	AssignerID            uuid.UUID                      `json:"assignerID"`
	AssigneeID            uuid.UUID                      `json:"assigneeID"`
	AppID                 uuid.UUID                      `json:"appID"`
	CapabilityName        string                         `json:"capabilityName"`
	CapabilityValue       string                         `json:"capabilityValue"`
	CapabilityGrantPolicy CapabilityAttributeBasedPolicy `json:"capabilityGrantPolicy,omitempty"`
	AuthorizeCapabilityID uuid.UUID                      `json:"authorizeCapabilityID"`
	CapabilitySignature   CapabilitySignature            `json:"capabilitySignature"`
	GrantType             string                         `json:"grantType,omitempty"`
	GrantCondition        string                         `json:"grantCondition,omitempty"`
}

// CapabilityAttributeBasedPolicy is a condition for Capability
type CapabilityAttributeBasedPolicy struct {
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
