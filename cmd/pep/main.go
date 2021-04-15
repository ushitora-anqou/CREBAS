package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/naoki9911/gofc"
	"github.com/vishvananda/netlink"
)

var apps = app.AppCollection{}
var pkgs = pkg.PkgCollection{}
var devices = app.DeviceCollection{}
var aclOfs = &ofswitch.OFSwitch{}
var extOfs = &ofswitch.OFSwitch{}
var appAddrPool = &ofswitch.IP4AddrPool{}
var extAddrPool = &ofswitch.IP4AddrPool{}
var controller = gofc.NewOFController()
var dnsServer = "8.8.8.8:53"

func main() {
	err := prepareNetwork()
	if err != nil {
		panic(err)
	}
	defer clearNetwork()
	startOFController()
	appendOFSwitchToController(aclOfs)
	waitOFSwitchConnectedToController(aclOfs)
	prepareTestPkg()
	go startDNSServer()
	StartAPIServer()
}

func startOFController() {
	log.Printf("Starting OpenFlow Controller...")
	go controller.ServerLoop(gofc.DEFAULT_PORT)
	log.Printf("Started OpenFlow Controller")
}

func appendOFSwitchToController(c *ofswitch.OFSwitch) {
	gofc.GetAppManager().RegistApplication(c)
}

func waitOFSwitchConnectedToController(c *ofswitch.OFSwitch) {
	for {
		if c.IsConnectedToController() {
			fmt.Println(c.Name + " is Connected!")
			break
		}
		time.Sleep(1 * time.Second)
	}

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

	err = aclOfs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		return err
	}

	extOfs = ofswitch.NewOFSwitch("crebas-ext-ofs")
	extOfs.Delete()
	err = extOfs.Create()
	if err != nil {
		return err
	}
	addr, err = netlink.ParseAddr("192.168.20.1/24")
	if err != nil {
		return err
	}
	extAddrPool = ofswitch.NewIP4AddrPool(addr)
	err = extAddrPool.LeaseWithAddr(addr)
	if err != nil {
		return err
	}

	err = extOfs.SetController("tcp:127.0.0.1:6653")
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
