package app

import (
	"os"
	"syscall"
	"testing"

	"github.com/vishvananda/netns"
)

func TestCreateVethPeer(t *testing.T) {
	err := createVethPeer("test-peer-a", "test-peer-b")
	if err != nil {
		t.Fatalf("Failed to create veth peer %#v", err)
	}
}

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

	_, err = netns.GetFromName(p.namespace)
	if err == nil {
		t.Fatalf("netns %v exists", p.namespace)
	}
}
