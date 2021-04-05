package main

import (
	"os/exec"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/vishvananda/netlink"
)

var apps = app.AppCollection{}
var pkgs = pkg.PkgCollection{}
var aclOfs = &ofswitch.OFSwitch{}
var appAddrPool = &ofswitch.IP4AddrPool{}

func main() {
	prepareNetwork()
	defer clearNetwork()
	prepareTestPkg()
	StartAPIServer()
}

func prepareNetwork() error {
	aclOfs = ofswitch.NewOFSwitch("crebas-acl-ofs")
	aclOfs.Delete()
	err := aclOfs.Create()
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr("192.168.10.1/24")
	if err != nil {
		return err
	}
	err = aclOfs.SetAddr(addr)
	if err != nil {
		return err
	}
	appAddrPool = ofswitch.NewIP4AddrPool(addr)
	err = appAddrPool.LeaseWithAddr(addr)
	if err != nil {
		return err
	}

	return nil
}

func clearNetwork() error {
	return aclOfs.Delete()
}

func prepareTestPkg() {
	testPkgsDir := "/tmp/pep_test"
	exec.Command("rm", "-rf", testPkgsDir).Run()
	exec.Command("mkdir", "-p", testPkgsDir).Run()

	pkgInfo := pkg.CreateSkeltonPackageInfo()
	pkgInfo.MetaInfo.CMD = []string{"echo", "HELLO"}
	err := pkg.CreatePackage(pkgInfo, testPkgsDir)
	if err != nil {
		panic(err)
	}

	err = pkgs.LoadPkgs(testPkgsDir)
	if err != nil {
		panic(err)
	}
}
