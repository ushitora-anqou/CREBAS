package app

import (
	"github.com/google/uuid"
)

// AppInterface is interface for application
// +gen * slice:"Where"
type AppInterface interface {
	Start() error
	Stop() error
	ID() uuid.UUID
	GetAppInfo() *AppInfo
}

type AppInfo struct {
	Id uuid.UUID `json:"id"`
}
