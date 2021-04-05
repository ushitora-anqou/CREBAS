package app

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func TestStartAndStopProcess(t *testing.T) {
	p, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}
	p.cmd = []string{"/usr/bin/sleep", "10"}
	err = p.Start()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	if !p.IsRunning() {
		t.Fatalf("Failed proc %v is not running", p.pid)
	}

	err = p.Stop()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	if p.IsRunning() {
		t.Fatalf("pid %v exists", p.pid)
	}

	for _, link := range p.links.Where(func(l *netlinkext.LinkExt) bool { return true }) {
		peerLinkName := link.GetLink().(*netlink.Veth).PeerName
		_, err = netlink.LinkByName(peerLinkName)
		if err == nil {
			t.Fatalf("veth link %v exists", peerLinkName)
		}
	}

	_, err = netns.GetFromName(p.namespace)
	if err == nil {
		t.Fatalf("netns %v exists", p.namespace)
	}
}

func TestLinkAttach(t *testing.T) {
	ofs := ofswitch.NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer ofs.Delete()

	ovsAddr, err := netlink.ParseAddr("192.168.100.1/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = ofs.SetAddr(ovsAddr)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	p, err := NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	procAddr, err := netlink.ParseAddr("192.168.100.2/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	_, err = p.AddLinkWithAddr(ofs, procAddr)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	p.cmd = []string{"/usr/bin/bash", "-c", "while true; do sleep 1; done"}
	err = p.Start()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	cmd := exec.Command("ping", "192.168.100.2", "-c", "1")
	out, err := cmd.Output()
	fmt.Println(string(out))
	if err != nil {
		t.Fatalf("Failed %v %v", err, string(out))
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		t.Fatalf("Failed exit code:%v", exitCode)
	}

	err = p.Stop()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}
