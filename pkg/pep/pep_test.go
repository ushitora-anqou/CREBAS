package main

import (
	"testing"
	"time"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/vishvananda/netlink"
)

func TestPeerCommunication(t *testing.T) {
	ovsName := "ovs-test-set"
	ofs := ofswitch.NewOFSwitch(ovsName)
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

	err = ofs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addrPool := ofswitch.NewIP4AddrPool(addr)
	err = addrPool.LeaseWithAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	startOFController(ofs)
	defer controller.Stop()

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
	link1, err := proc1.AddLinkWithAddr(ofs, proc1Addr)
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
	link2, err := proc2.AddLinkWithAddr(ofs, proc2Addr)
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

	time.Sleep(5 * time.Second)
	if proc2.GetExitCode() != 0 {
		t.Fatalf("proc2 exit code: %v", proc2.GetExitCode())
	}

	proc1.Stop()
	proc2.Stop()
}
