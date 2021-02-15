package app

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func TestStartAndStopProcess(t *testing.T) {
	p := NewLinuxProcess()
	err := p.Create()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}
	p.cmd = []string{"/usr/bin/sleep", "1"}
	err = p.Start()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = p.Stop()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = p.Delete()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	proc, err := os.FindProcess(p.pid)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
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

	p := NewLinuxProcess()
	err = p.Create()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}
	defer p.Delete()

	procAddr, err := netlink.ParseAddr("192.168.100.2/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	link, err := p.AddLink(ofs)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	link.SetAddr(procAddr)

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
