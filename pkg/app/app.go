package app

import (
	"encoding/json"
	"io/ioutil"

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

type PackageInfo struct {
	MetaInfo PackageMetaInfo
}

type PackageMetaInfo struct {
	Name     string
	PkgID    uuid.UUID
	VendorID uuid.UUID
	CMD      string
}

// OpenPackageInfo read pkgInfo and return PackageInfo
func OpenPackageInfo(pkgInfoPath string) (*PackageInfo, error) {
	bytes, err := ioutil.ReadFile(pkgInfoPath)
	if err != nil {
		return nil, err
	}
	var pkgInfo PackageInfo
	err = json.Unmarshal(bytes, &pkgInfo)
	if err != nil {
		return nil, err
	}

	return &pkgInfo, nil
}

// Verify PackageInfo
func (p *PackageInfo) Verify() bool {
	return true
}
