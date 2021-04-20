package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestPeerCommunication(t *testing.T) {
	startOFController()
	defer controller.Stop()
	ovsName := "ovs-test-set"
	ofs := ofswitch.NewOFSwitch(ovsName)
	ofs.Delete()
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer ofs.Delete()
	fmt.Println(ofs.DatapathID)

	addr, err := netlink.ParseAddr("192.168.10.1/24")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.SetAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	appendOFSwitchToController(ofs)

	err = ofs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addrPool := ofswitch.NewIP4AddrPool(addr)
	err = addrPool.LeaseWithAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	waitOFSwitchConnectedToController(ofs)

	pkg1 := pkg.CreateSkeltonPackageInfo()
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 5"}
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc1Addr, err := addrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	link1, err := proc1.AddLinkWithAddr(ofs, netlinkext.ACLOFSwitch, proc1Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	pkg2 := pkg.CreateSkeltonPackageInfo()
	pkg2.MetaInfo.CMD = []string{"/bin/bash", "-c", "ping -c 1 192.168.10.2"}
	proc2, err := app.NewLinuxProcessFromPkgInfo(pkg2)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc2Addr, err := addrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	link2, err := proc2.AddLinkWithAddr(ofs, netlinkext.ACLOFSwitch, proc2Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	err = ofs.AddARPFlow(link1, link2)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	err = ofs.AddICMPFlow(link1, link2)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	proc1.Start()
	time.Sleep(100 * time.Millisecond)
	proc2.Start()

	time.Sleep(2 * time.Second)
	if proc2.GetExitCode() != 0 {
		t.Fatalf("proc2 exit code: %v", proc2.GetExitCode())
	}

	proc1.Stop()
	proc2.Stop()
}

func TestPeerCommunication2(t *testing.T) {
	startOFController()
	defer controller.Stop()
	time.Sleep(time.Second)
	pkgDir := "/tmp/pep_test"
	ovsName := "ovs-test-set"
	ofs := ofswitch.NewOFSwitch(ovsName)
	ofs.Delete()
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer ofs.Delete()

	addr, err := netlink.ParseAddr("192.168.10.1/24")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.SetAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	appendOFSwitchToController(ofs)

	err = ofs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addrPool := ofswitch.NewIP4AddrPool(addr)
	err = addrPool.LeaseWithAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	waitOFSwitchConnectedToController(ofs)

	pkg1 := pkg.CreateSkeltonPackageInfo()
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 5"}
	err = pkg.CreateUnpackedPackage(pkg1, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc1Addr, err := addrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	_, err = proc1.AddLinkWithAddr(ofs, netlinkext.ACLOFSwitch, proc1Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	pkg2 := pkg.CreateSkeltonPackageInfo()
	pkg2.MetaInfo.CMD = []string{"/bin/bash", "-c", "ping -c 1 -W 1 192.168.10.2"}
	err = pkg.CreateUnpackedPackage(pkg2, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	proc2, err := app.NewLinuxProcessFromPkgInfo(pkg2)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc2Addr, err := addrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	_, err = proc2.AddLinkWithAddr(ofs, netlinkext.ACLOFSwitch, proc2Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	proc1.Start()
	time.Sleep(100 * time.Millisecond)
	proc2.Start()

	time.Sleep(5 * time.Second)
	if proc2.GetExitCode() == 0 {
		t.Fatalf("proc2 exit code: %v", proc2.GetExitCode())
	}

	proc1.Stop()
	proc2.Stop()
}

func TestExtCommunication(t *testing.T) {
	pkgDir := "/tmp/pep_test"
	ovsName := "ovs-test-set"
	apps.Clear()

	startOFController()
	defer controller.Stop()
	time.Sleep(time.Second)

	ofs := ofswitch.NewOFSwitch(ovsName)
	ofs.Delete()
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer ofs.Delete()

	appendOFSwitchToController(ofs)

	addr, err := netlink.ParseAddr("192.168.10.1/24")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.SetAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	err = ofs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addrPool := ofswitch.NewIP4AddrPool(addr)
	err = addrPool.LeaseWithAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	ovsName2 := "ovs-test-set2"
	ofs2 := ofswitch.NewOFSwitch(ovsName2)
	err = ofs2.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer ofs2.Delete()

	appendOFSwitchToController(ofs2)

	extAddr, err := netlink.ParseAddr("192.168.20.1/24")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	err = ofs2.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addrPool2 := ofswitch.NewIP4AddrPool(extAddr)
	err = addrPool2.LeaseWithAddr(extAddr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	waitOFSwitchConnectedToController(ofs)
	waitOFSwitchConnectedToController(ofs2)

	pkg1 := pkg.CreateSkeltonPackageInfo()
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 10"}
	err = pkg.CreateUnpackedPackage(pkg1, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc1Addr, err := addrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc1Link, err := proc1.AddLinkWithAddr(ofs, netlinkext.ACLOFSwitch, proc1Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = proc1.SetDefaultRoute(addr.IP)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	err = ofs.AddHostRestrictedFlow(proc1Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	err = ofs.AddHostUnicastTCPDstFlow(proc1Link, 8080)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	proc1Link2, err := proc1.AddLinkWithAddr(ofs2, netlinkext.ExternalOFSwitch, extAddr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	apps.Add(proc1)

	apiUrl := "http://192.168.10.1:8080/app/" + proc1.ID().String()
	pkg2 := pkg.CreateSkeltonPackageInfo()
	pkg2.MetaInfo.CMD = []string{"/bin/bash", "-c", "curl " + apiUrl}
	proc2, err := app.NewLinuxProcessFromPkgInfo(pkg2)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc2Addr, err := addrPool2.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc2Link, err := proc2.AddLinkWithAddr(ofs2, netlinkext.ExternalOFSwitch, proc2Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = proc2.SetDefaultRoute(extAddr.IP)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	ofs2.AddTunnelFlow(proc1Link2, proc2Link)

	proc1.Start()
	time.Sleep(4 * time.Second)
	proc2.Start()

	time.Sleep(5 * time.Second)
	if proc2.GetExitCode() != 0 {
		t.Fatalf("proc2 exit code: %v", proc2.GetExitCode())
	}

	proc1.Stop()
	proc2.Stop()
}

func TestDNSCommunication(t *testing.T) {
	pkgDir := "/tmp/pep_test"
	apps.Clear()

	startOFController()
	defer controller.Stop()
	time.Sleep(time.Second)

	err := prepareNetwork()
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	defer clearNetwork()

	go startDNSServer(aclOfs)

	pkg1 := pkg.CreateSkeltonPackageInfo()
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "dig net.ist.i.kyoto-u.ac.jp"}
	err = pkg.CreateUnpackedPackage(pkg1, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc1Addr, err := appAddrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	proc1Link, err := proc1.AddLinkWithAddr(aclOfs, netlinkext.ACLOFSwitch, proc1Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = proc1.SetDNSServer(aclOfs.Link.Addr.IP)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = aclOfs.AddHostRestrictedFlow(proc1Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = aclOfs.AddHostUnicastUDPDstFlow(proc1Link, 53)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	cap := capability.NewCreateSkeltonCapability()
	cap.CapabilityName = capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	cap.CapabilityValue = "*.net.ist.i.kyoto-u.ac.jp"
	proc1.Capabilities().Add(cap)

	apps.Add(proc1)
	proc1.Start()
	time.Sleep(5 * time.Second)
	assert.Equal(t, proc1.IsRunning(), false)
	assert.Equal(t, proc1.GetExitCode(), 0)

	proc1.Stop()
}

func TestDHCPCommunication(t *testing.T) {
	pkgDir := "/tmp/pep_test"
	apps.Clear()

	startOFController()
	defer controller.Stop()
	time.Sleep(time.Second)

	err := prepareNetwork()
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	defer clearNetwork()
	go StartDHCPServer()

	pkg1 := pkg.CreateSkeltonPackageInfo()
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	proc1Link, err := proc1.AddLink(extOfs, netlinkext.ExternalOFSwitch)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	peerName := proc1Link.GetLink().Attrs().Name
	err = extOfs.AddHostARPFlow(proc1Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = extOfs.AddHostDHCPFlow(proc1Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "dhclient " + peerName + "&& ping -c 1 192.168.20.1"}
	pkg1.TestUse = true
	err = pkg.CreateUnpackedPackage(pkg1, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	pkg2 := pkg.CreateSkeltonPackageInfo()
	pkg2.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 500"}
	proc2, err := app.NewLinuxProcessFromPkgInfo(pkg2)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = pkg.CreateUnpackedPackage(pkg2, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	deviceIP, err := extAddrPool.Lease()
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	device := app.Device{
		HWAddress: proc1Link.GetLink().Attrs().HardwareAddr,
		IPAddress: deviceIP,
		App:       proc2,
		OfPort:    proc1Link.Ofport,
		ViaWlan:   false,
	}
	fmt.Println(proc1.NameSpace())
	devices.Add(&device)
	//apps.Add(proc1)
	proc1.Start()
	time.Sleep(10 * time.Second)
	assert.Equal(t, proc1.IsRunning(), false)
	assert.Equal(t, proc1.GetExitCode(), 0)

	proc1.Stop()
	proc2.Stop()
}

func TestReflection(t *testing.T) {
	pkgDir := "/tmp/pep_test"
	apps.Clear()

	startOFController()
	defer controller.Stop()
	time.Sleep(time.Second)

	err := prepareNetwork()
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	defer clearNetwork()

	app1 := pkg.CreateSkeltonPackageInfo()
	procApp1, err := app.NewLinuxProcessFromPkgInfo(app1)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	procApp1Link, err := procApp1.AddLink(extOfs, netlinkext.ExternalOFSwitch)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	app1.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 500"}
	app1.TestUse = false
	err = pkg.CreateUnpackedPackage(app1, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	app2 := pkg.CreateSkeltonPackageInfo()
	procApp2, err := app.NewLinuxProcessFromPkgInfo(app2)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	procApp2Link, err := procApp2.AddLink(extOfs, netlinkext.ExternalOFSwitch)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	app2.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 500"}
	app2.TestUse = false
	err = pkg.CreateUnpackedPackage(app2, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	device1 := pkg.CreateSkeltonPackageInfo()
	device1.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 500"}
	device1.TestUse = true
	procDevice1, err := app.NewLinuxProcessFromPkgInfo(device1)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	device1Addr, err := netlink.ParseAddr("192.168.20.2/24")
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	procDevice1Link, err := procDevice1.AddLinkWithAddr(extOfs, netlinkext.ExternalOFSwitch, device1Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = pkg.CreateUnpackedPackage(device1, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = extOfs.AddDeviceAppTunnelFlow(procDevice1Link, procApp1Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	device2 := pkg.CreateSkeltonPackageInfo()
	device2.MetaInfo.CMD = []string{"/bin/bash", "-c", "ping 192.168.20.2 -c 1"}
	device2.TestUse = true
	procDevice2, err := app.NewLinuxProcessFromPkgInfo(device2)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	device2Addr, err := netlink.ParseAddr("192.168.20.3/24")
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	procDevice2Link, err := procDevice2.AddLinkWithAddr(extOfs, netlinkext.ExternalOFSwitch, device2Addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = pkg.CreateUnpackedPackage(device2, pkgDir)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = extOfs.AddDeviceAppTunnelFlow(procDevice2Link, procApp2Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}

	err = extOfs.AddAppsARPFlow(procDevice1Link, procApp1Link, procDevice2Link, procApp2Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	err = extOfs.AddAppsICMPFlow(procDevice1Link, procApp1Link, procDevice2Link, procApp2Link)
	if err != nil {
		t.Fatalf("failed test %v", err)
	}
	fmt.Println(procDevice2Link.GetLink().Attrs().Name)
	fmt.Println(procDevice2.NameSpace())

	procApp1.Start()
	defer procApp1.Stop()
	procApp2.Start()
	defer procApp2.Stop()
	procDevice1.Start()
	defer procDevice1.Stop()
	procDevice2.Start()
	defer procDevice2.Stop()

	time.Sleep(10 * time.Second)

	assert.Equal(t, procDevice2.IsRunning(), false)
	assert.Equal(t, procDevice2.GetExitCode(), 0)

}
func init() {
	go StartAPIServer()
}
