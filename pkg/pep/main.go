package main

import (
	"os/exec"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/pkg"
)

var apps = app.AppCollection{}
var pkgs = pkg.PkgCollection{}

func main() {
	prepareTestPkg()
	StartAPIServer()
}

func prepareTestPkg() {
	testPkgsDir := "/tmp/pep_test"
	exec.Command("rm", "-rf", testPkgsDir).Run()
	exec.Command("mkdir", "-p", testPkgsDir).Run()

	pkgInfo := pkg.CreateSkeltonPackageInfo()
	pkgInfo.MetaInfo.CMD = "echo HELLO"
	err := pkg.CreatePackage(pkgInfo, testPkgsDir)
	if err != nil {
		panic(err)
	}

	err = pkgs.LoadPkgs(testPkgsDir)
	if err != nil {
		panic(err)
	}
}
