package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Printf("error: Invalid cmdline args")
		return
	}

	if flag.Args()[0] == "show" && len(flag.Args()) == 2 {
		showPkgInfo(flag.Args()[1])
		return
	}

	if flag.Args()[0] == "pack" && len(flag.Args()) == 2 {
		packPkg(flag.Args()[1])
		return
	}

	if flag.Args()[0] == "showExample" {
		showExample()
		return
	}

}

func packPkg(pkgDir string) {
	pkgDirAbs, err := filepath.Abs(pkgDir)
	if err != nil {
		panic(err)
	}

	pkgInfoPath := filepath.Join(pkgDirAbs, "pkgInfo.json")
	pkgInfo, err := app.OpenPackageInfo(pkgInfoPath)
	if err != nil {
		panic(err)
	}
	log.Printf("info: Load Package Info(%v)", pkgInfoPath)

	if !pkgInfo.Verify() {
		panic(fmt.Errorf("invalid package info"))
	}
	log.Printf("info: Successfully verified Package Info")

	packageName := pkgInfo.MetaInfo.Name + ".tar.gz"
	log.Printf("info: Packing application to %v", packageName)
	err = exec.Command("tar", "-zcvf", packageName, "-C", pkgDirAbs+"/", ".").Run()
	if err != nil {
		panic(err)
	}

	log.Printf("info: Successfully packed Package %v", packageName)
}

func showPkgInfo(packagePath string) *app.PackageInfo {
	pkgPathAbs, err := filepath.Abs(packagePath)
	if err != nil {
		panic(err)
	}
	packageExt := filepath.Ext(packagePath)
	tmpDir := ""
	if packageExt != ".json" {
		tmpDir = unpackPkg(pkgPathAbs)
		pkgPathAbs = filepath.Join(tmpDir, "pkgInfo.json")
	}

	pkgInfo, err := app.OpenPackageInfo(pkgPathAbs)
	if err != nil {
		panic(err)
	}

	if !pkgInfo.Verify() {
		panic(fmt.Errorf("invalid package info"))
	}

	fmt.Printf("MetaInfo:\n")
	fmt.Printf("- Name:     %v\n", pkgInfo.MetaInfo.Name)
	fmt.Printf("- PkgID:    %v\n", pkgInfo.MetaInfo.PkgID)
	fmt.Printf("- VendorID: %v\n", pkgInfo.MetaInfo.VendorID)
	fmt.Printf("- CMD:      %v\n", pkgInfo.MetaInfo.CMD)

	if packageExt != ".json" && tmpDir != "" {
		err = exec.Command("rm", "-rf", tmpDir).Run()
		if err != nil {
			panic(err)
		}
	}

	return pkgInfo
}

func unpackPkg(pkgPath string) string {
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

func showExample() {
	pkgID, _ := uuid.NewRandom()
	vendorID, _ := uuid.NewRandom()

	pkgInfo := app.PackageInfo{
		MetaInfo: app.PackageMetaInfo{
			Name:     "test-pkg",
			PkgID:    pkgID,
			VendorID: vendorID,
			CMD:      "ping 127.0.0.1",
		},
	}

	pkgInfoJson, _ := json.Marshal(pkgInfo)
	fmt.Println(string(pkgInfoJson))
}

func createSkeltonPackage(pkgPath string) {
	pkgID, _ := uuid.NewRandom()
	vendorID, _ := uuid.NewRandom()

	pkgInfo := app.PackageInfo{
		MetaInfo: app.PackageMetaInfo{
			Name:     "test-pkg",
			PkgID:    pkgID,
			VendorID: vendorID,
			CMD:      "ping 127.0.0.1",
		},
	}

	pkgInfoJson, _ := json.Marshal(pkgInfo)

	err := ioutil.WriteFile(filepath.Join(pkgPath, "pkgInfo.json"), pkgInfoJson, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
