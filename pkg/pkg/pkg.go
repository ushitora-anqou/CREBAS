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
	PkgPath         string
	UnpackedPkgPath string
	MetaInfo        PackageMetaInfo
}

type PackageMetaInfo struct {
	Name     string
	PkgID    uuid.UUID
	VendorID uuid.UUID
	CMD      []string
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

	pkgDir, err := filepath.Abs(filepath.Dir(pkgInfoPath))
	if err != nil {
		return nil, err
	}
	pkgInfo.UnpackedPkgPath = pkgDir

	return &pkgInfo, nil
}

// Verify PackageInfo
func (p *PackageInfo) Verify() bool {
	return true
}

func UnpackPkg(pkgPath string) (*PackageInfo, error) {
	pkgPathAbs, err := filepath.Abs(pkgPath)
	if err != nil {
		return nil, err
	}

	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	uuidStr := uuidObj.String()
	tmpDir := "/tmp/apppackager/" + uuidStr
	log.Printf("info: Creating temporary directory %v", tmpDir)
	err = exec.Command("mkdir", "-p", tmpDir).Run()
	if err != nil {
		return nil, err
	}
	prevDir, _ := filepath.Abs(".")
	os.Chdir(tmpDir)
	log.Printf("info: Unpacking %v", pkgPathAbs)
	err = exec.Command("tar", "-xvf", pkgPathAbs).Run()
	if err != nil {
		return nil, err
	}
	log.Printf("info: Unpacked package to %v", tmpDir)
	os.Chdir(prevDir)

	pkgInfo, err := OpenPackageInfo(filepath.Join(tmpDir, "pkgInfo.json"))
	if err != nil {
		return nil, err
	}
	pkgInfo.PkgPath = pkgPathAbs

	return pkgInfo, nil
}

func CreateSkeltonPackage(pkgPath string) *PackageInfo {
	pkgID, _ := uuid.NewRandom()
	vendorID, _ := uuid.NewRandom()

	pkgInfo := PackageInfo{
		MetaInfo: PackageMetaInfo{
			Name:     "test-pkg",
			PkgID:    pkgID,
			VendorID: vendorID,
			CMD:      []string{"ping", "127.0.0.1"},
		},
	}

	pkgInfoJson, _ := json.Marshal(pkgInfo)

	err := ioutil.WriteFile(filepath.Join(pkgPath, "pkgInfo.json"), pkgInfoJson, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return &pkgInfo
}

func CreateSkeltonUnpackedPackage(pkgRootDir string) *PackageInfo {
	pathID, _ := uuid.NewRandom()
	pkgPath := filepath.Join(pkgRootDir, pathID.String())
	exec.Command("mkdir", "-p", pkgPath)
	pkgInfo := CreateSkeltonPackage(pkgPath)
	pkgInfo.UnpackedPkgPath = pkgPath

	return pkgInfo
}

func CreateSkeltonPackageInfo() *PackageInfo {
	pkgID, _ := uuid.NewRandom()
	vendorID, _ := uuid.NewRandom()

	pkgInfo := PackageInfo{
		MetaInfo: PackageMetaInfo{
			Name:     "test-pkg",
			PkgID:    pkgID,
			VendorID: vendorID,
			CMD:      []string{"ping", "127.0.0.1"},
		},
	}

	return &pkgInfo
}

func CreateUnpackedPackage(pkgInfo *PackageInfo, pkgPath string) error {
	uuid, _ := uuid.NewRandom()
	tmpPkgDir := filepath.Join(pkgPath, uuid.String())
	err := exec.Command("mkdir", "-p", tmpPkgDir).Run()
	if err != nil {
		return err
	}

	pkgInfoJson, _ := json.Marshal(pkgInfo)
	err = ioutil.WriteFile(filepath.Join(tmpPkgDir, "pkgInfo.json"), pkgInfoJson, os.ModePerm)
	if err != nil {
		return err
	}

	pkgInfo.UnpackedPkgPath = tmpPkgDir
	return nil
}

func CreatePackage(pkgInfo *PackageInfo, pkgPath string) error {
	uuid, _ := uuid.NewRandom()
	tmpPkgDir := filepath.Join(pkgPath, uuid.String())
	err := exec.Command("mkdir", "-p", tmpPkgDir).Run()
	if err != nil {
		return err
	}

	pkgInfoJson, _ := json.Marshal(pkgInfo)
	err = ioutil.WriteFile(filepath.Join(tmpPkgDir, "pkgInfo.json"), pkgInfoJson, os.ModePerm)
	if err != nil {
		return err
	}

	packageName := pkgInfo.MetaInfo.Name + ".tar.gz"
	log.Printf("info: Packing application to %v", packageName)
	err = exec.Command("tar", "-zcvf", filepath.Join(pkgPath, packageName), "-C", tmpPkgDir, ".").Run()
	if err != nil {
		return err
	}

	err = exec.Command("rm", "-rf", tmpPkgDir).Run()
	if err != nil {
		return err
	}

	return nil
}
