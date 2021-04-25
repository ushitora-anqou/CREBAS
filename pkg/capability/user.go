package capability

import "github.com/google/uuid"

type UserGrantPolicy struct {
	UserGrantPolicyID uuid.UUID `json:"userGrantPolicyID"`
	CapabilityID      uuid.UUID `json:"capabilityID"`
	Grant             bool      `json:"grant"`
	RequesterID       uuid.UUID `json:"targetAppID"`
}
