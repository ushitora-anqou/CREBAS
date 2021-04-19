package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
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
var pepConfig = NewConfig()

func main() {
	startOFController()
	err := prepareNetwork()
	if err != nil {
		panic(err)
	}
	defer clearNetwork()
	err = setupWiFi()
	if err != nil {
		panic(err)
	}
	err = prepareTestPkg()
	if err != nil {
		panic(err)
	}
	go startDNSServer(aclOfs)
	go StartDHCPServer()
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
	aclOfs = ofswitch.NewOFSwitch(pepConfig.aclOfsName)
	aclOfs.Delete()
	err := aclOfs.Create()
	if err != nil {
		return err
	}

	appendOFSwitchToController(aclOfs)

	addr, err := netlink.ParseAddr(pepConfig.aclOfsAddr)
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

	waitOFSwitchConnectedToController(aclOfs)

	extOfs = ofswitch.NewOFSwitch(pepConfig.extOfsName)
	extOfs.Delete()
	err = extOfs.Create()
	if err != nil {
		return err
	}

	appendOFSwitchToController(extOfs)

	addr, err = netlink.ParseAddr(pepConfig.extOfsAddr)
	if err != nil {
		return err
	}
	err = extOfs.SetAddr(addr)
	if err != nil {
		return err
	}
	extAddrPool = ofswitch.NewIP4AddrPool(addr)
	err = extAddrPool.LeaseWithAddr(addr)
	if err != nil {
		return err
	}

	extAppAddr, err := netlink.ParseAddr(pepConfig.extOfsAppAddr)
	if err != nil {
		return err
	}
	err = extAddrPool.LeaseWithAddr(extAppAddr)
	if err != nil {
		return err
	}

	err = extOfs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		return err
	}

	waitOFSwitchConnectedToController(extOfs)

	return nil
}

func clearNetwork() error {
	err := aclOfs.Delete()
	if err != nil {
		return err
	}
	err = extOfs.Delete()
	if err != nil {
		return err
	}

	return nil
}

func prepareTestPkg() error {
	pkgDir := "/tmp/pep_test"

	pkg1 := pkg.CreateSkeltonPackageInfo()
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 500"}
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		return err
	}
	err = pkg.CreateUnpackedPackage(pkg1, pkgDir)
	if err != nil {
		return err
	}

	deviceIP, err := extAddrPool.Lease()
	if err != nil {
		return err
	}
	hwAddr, err := net.ParseMAC("58:cb:52:56:73:21")
	if err != nil {
		return err
	}
	device := &app.Device{
		HWAddress: hwAddr,
		IPAddress: deviceIP,
		App:       proc1,
		OfPort:    pepConfig.wifiLink.GetOfPort(),
		ViaWlan:   true,
	}

	devices.Add(device)

	return nil
}

func setupWiFi() error {
	link, err := netlink.LinkByName("wlan0")
	if err != nil {
		return err
	}
	linkExt := &netlinkext.LinkExt{}
	linkExt.SetLink(link)

	for {
		ofport, err := ofswitch.GetOFPortByLinkName("wlan0")
		if err == nil {
			log.Printf("Found OfPort wlan0 %v", ofport)
			linkExt.Ofport = ofport
			break
		}
		log.Printf("Not Found OfPort for wlan0")
		time.Sleep(1 * time.Second)
	}
	err = extOfs.ResetController()
	if err != nil {
		return err
	}
	log.Printf("Reset Controller")
	waitOFSwitchConnectedToController(extOfs)

	pepConfig.wifiLink = linkExt

	err = extOfs.AddHostEAPoLFlow(linkExt)
	if err != nil {
		return err
	}
	err = extOfs.AddHostAggregatedARPFlow(linkExt)
	if err != nil {
		return err
	}
	err = extOfs.AddHostAggregatedDHCPFlow(linkExt)
	if err != nil {
		return err
	}
	return nil
}
