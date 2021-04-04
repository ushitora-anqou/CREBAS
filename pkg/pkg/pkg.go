package pkg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

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

func UnpackPkg(pkgPath string) string {
	pkgPathAbs, err := filepath.Abs(pkgPath)
	if err != nil {
		panic(err)
	}

	uuidObj, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	uuidStr := uuidObj.String()
	tmpDir := "/tmp/apppackager/" + uuidStr
	log.Printf("info: Creating temporary directory %v", tmpDir)
	err = exec.Command("mkdir", "-p", tmpDir).Run()
	if err != nil {
		panic(err)
	}
	prevDir, _ := filepath.Abs(".")
	os.Chdir(tmpDir)
	log.Printf("info: Unpacking %v", pkgPathAbs)
	err = exec.Command("tar", "-xvf", pkgPathAbs).Run()
	if err != nil {
		panic(err)
	}
	log.Printf("info: Unpacked package to %v", tmpDir)
	os.Chdir(prevDir)

	return tmpDir
}
