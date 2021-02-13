package netlinkext

import (
	"testing"

	"github.com/vishvananda/netns"
)

func TestCreateNetns(t *testing.T) {
	netnsName := "test-netns"
	_, err := CreateNetns(netnsName)
	if err != nil {
		t.Fatalf("Failed %#v", err)
	}

	_, err = netns.GetFromName("test-netns-2")
	if err == nil {
		t.Fatalf("Failed %#v", err)
	}

	_, err = netns.GetFromName(netnsName)
	if err != nil {
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
