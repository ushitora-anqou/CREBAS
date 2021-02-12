package netlinkext

import (
	"testing"

	"github.com/vishvananda/netlink"
)

func createLinkExt(name string) *LinkExt {
	veth := netlink.Veth{}
	veth.LinkAttrs = netlink.NewLinkAttrs()
	veth.LinkAttrs.Name = name
	linkExt := &LinkExt{}
	linkExt.link = &veth

	return linkExt
}

func TestAdd(t *testing.T) {
	linkExt := createLinkExt("test")
	linkCollection := NewLinkCollection()
	linkCollection.Add(linkExt)

	linkTest := linkCollection.GetByIndex(0)
	if linkExt != linkTest {
		t.Fatalf("Failed")
	}
}

func TestRemove(t *testing.T) {
	linkExt := createLinkExt("test-1")
	linkExt2 := createLinkExt("test-2")

	linkCollection := NewLinkCollection()
	linkCollection.Add(linkExt)
	linkCollection.Add(linkExt2)

	if count := linkCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	err := linkCollection.Remove(linkExt)
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	linkTest := linkCollection.GetByIndex(0)

	if linkTest != linkExt2 {
		t.Fatalf("Failed")
	}

	if linkTest == linkExt {
		t.Fatalf("Failed")
	}
}

func TestWhere(t *testing.T) {
	linkExt := createLinkExt("test-1")
	linkExt2 := createLinkExt("test-2")

	linkCollection := NewLinkCollection()
	linkCollection.Add(linkExt)
	linkCollection.Add(linkExt2)

	if count := linkCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	slices := linkCollection.Where(func(link *LinkExt) bool {
		return link.link.Attrs().Name == "test-1"
	})

	if slices[0] == linkExt2 {
		t.Fatalf("Failed")
	}

	if slices[0] != linkExt {
		t.Fatalf("Failed")
	}
}
