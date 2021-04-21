package app

import (
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
)

// AppInterface is interface for application
// +gen * slice:"Where"
type AppInterface interface {
	Start() error
	Stop() error
	ID() uuid.UUID
	GetAppInfo() *AppInfo
	IsRunning() bool
	GetExitCode() int
	SetDevice(*Device) error
	GetDevice() *Device
	Capabilities() *capability.CapabilityCollection
	Links() *netlinkext.LinkCollection
}

type AppInfo struct {
	Id                      uuid.UUID `json:"id"`
	ACLLinkName             string    `json:"aclLinkName"`
	ACLLinkPeerHWAddress    string    `json:"aclLinkPeerHWAddress`
	DeviceLinkName          string    `json:"deviceLinkName"`
	DeviceLinkPeerHWAddress string    `json:"deviceLinkPeerHWAddress`
	OvsACLHWAddr            string    `json:"ovsACLHWAddr"`
	OvsExtHWAddr            string    `json:"ovsExtHWAddr"`
}
