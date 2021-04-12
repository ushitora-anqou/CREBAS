package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/pkg"
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
	pkgInfo, err := pkg.OpenPackageInfo(pkgInfoPath)
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
	err = exec.Command("tar", "-zcvf", packageName, "-C", pkgDirAbs, ".").Run()
	if err != nil {
		panic(err)
	}

	log.Printf("info: Successfully packed Package %v", packageName)
}

func showPkgInfo(packagePath string) *pkg.PackageInfo {
	pkgPathAbs, err := filepath.Abs(packagePath)
	if err != nil {
		panic(err)
	}
	if !strings.HasSuffix(pkgPathAbs, ".tar.gz") {
		panic(fmt.Errorf("Unexpected file %v", pkgPathAbs))
	}
	pkgInfo, err := pkg.UnpackPkg(pkgPathAbs)
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

	err = exec.Command("rm", "-rf", pkgInfo.UnpackedPkgPath).Run()
	if err != nil {
		panic(err)
	}

	return pkgInfo
}

func showExample() {
	pkgID, _ := uuid.NewRandom()
	vendorID, _ := uuid.NewRandom()

	pkgInfo := pkg.PackageInfo{
		MetaInfo: pkg.PackageMetaInfo{
			Name:     "test-pkg",
			PkgID:    pkgID,
			VendorID: vendorID,
			CMD:      []string{"ping", "127.0.0.1"},
		},
	}

	pkgInfoJson, _ := json.Marshal(pkgInfo)
	fmt.Println(string(pkgInfoJson))
}
