package netlinkext

import (
	"testing"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func TestConfigureVethWithNetns(t *testing.T) {
	linkName := "veth-test"
	peerName := "veth-test-peer"
	netnsName := "netns-test"

	linkExt := NewLinkExtVeth(linkName, peerName)

	handle, err := CreateNetns(netnsName)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}
	defer handle.Close()

	err = linkExt.Create()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	err = linkExt.SetNsByName(netnsName)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	_, err = netlink.LinkByName(peerName)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	_, err = GetLinkByNameWithNetns(linkName, handle)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	addr, _ := netlink.ParseAddr("192.168.100.2/32")
	badAddr, _ := netlink.ParseAddr("192.168.100.3/32")
	err = linkExt.SetAddr(addr)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	addrs, err := GetLinkAddrWithNetns(linkExt.link, netlink.FAMILY_V4, handle)
	if !addrs[0].Equal(*addr) {
		t.Fatalf("Failed expected %v actual %v", addr, addrs[0])
	}
	if addrs[0].Equal(*badAddr) {
		t.Fatalf("Failed not expected %v actual %v", badAddr, addrs[0])
	}

	err = linkExt.Delete()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	_, err = GetLinkByNameWithNetns(linkName, handle)
	if err == nil {
		t.Fatalf("Failed %#v", err)
	}

	err = netns.DeleteNamed(netnsName)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	_, err = netns.GetFromName(netnsName)
	if err == nil {
		t.Fatalf("Failed %#v", err)
	}
}
