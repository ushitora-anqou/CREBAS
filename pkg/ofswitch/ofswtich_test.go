package ofswitch

import (
	"testing"

	"github.com/digitalocean/go-openvswitch/ovs"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func ovsBridgeExists(bridgeName string) (bool, error) {
	client := ovs.New()
	bridges, err := client.VSwitch.ListBridges()
	if err != nil {
		return false, err
	}

	for _, bridge := range bridges {
		if bridge == bridgeName {
			return true, nil
		}
	}

	return false, nil
}

func getOvsController(bridgeName string) (string, error) {
	client := ovs.New()
	return client.VSwitch.GetController(bridgeName)
}

func TestCreate(t *testing.T) {
	ofs := NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	linkName := ofs.link.GetLink().Attrs().Name
	if linkName != ofs.Name {
		t.Fatalf("failed test expected %v actual %v", ofs.Name, linkName)
	}

	exist, err := ovsBridgeExists(ofs.Name)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if !exist {
		t.Fatalf("failed test ovs bridge %v does not exist", ofs.Name)
	}

	if ofs.DatapathID == 0 {
		t.Fatalf("invalid datapathID")
	}
}

func TestCreateAndRemove(t *testing.T) {
	ofs := NewOFSwitch("ovs-test-hoge")
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.Delete()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	exist, err := ovsBridgeExists(ofs.Name)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if exist {
		t.Fatalf("failed test ovs bridge %v remained", ofs.Name)
	}
}

func TestSetController(t *testing.T) {
	ovsName := "ovs-test-set"
	ofs := NewOFSwitch(ovsName)
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.SetController("localhost:6655")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	controllerURL, err := getOvsController(ovsName)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if controllerURL != "localhost:6655" {
		t.Fatalf("failed to test expected localhost:6655 actual %v", controllerURL)
	}

	err = ofs.Delete()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
}

func TestSetAddressV4(t *testing.T) {
	ovsName := "ovs-test-set"
	ofs := NewOFSwitch(ovsName)
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addr, err := netlink.ParseAddr("192.168.10.1/24")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = ofs.SetAddr(addr)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	link, err := netlink.LinkByName(ofs.Name)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	err = ofs.Delete()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	for _, ad := range addrs {
		if ad.Equal(*addr) {
			return
		}
	}

	t.Fatalf("failed test %#v does not found", addr.String())
}

func TestAttachLink(t *testing.T) {
	linkName := "veth-test"
	peerName := "veth-test-peer"
	netnsName := "netns-test"
	ovsName := "ovs-test-set"

	ofs := NewOFSwitch(ovsName)
	err := ofs.Create()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	linkExt := netlinkext.NewLinkExtVeth(linkName, peerName)
	handle, err := netlinkext.CreateNetns(netnsName)
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

	err = ofs.AttachLink(linkExt, netlinkext.ACLOFSwitch)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	if linkExt.Ofport != 1 {
		t.Fatalf("Failed expected:1 actual:%v", linkExt.Ofport)
	}

	err = linkExt.Delete()
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	err = netns.DeleteNamed(netnsName)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	err = ofs.Delete()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
}
