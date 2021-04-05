package ofswitch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestLeaseWithAddr(t *testing.T) {
	subnet, err := netlink.ParseAddr("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	ipPool := NewIP4AddrPool(subnet)

	ip, err := netlink.ParseAddr("192.168.1.1/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = ipPool.LeaseWithAddr(ip)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = ipPool.LeaseWithAddr(ip)
	if err == nil {
		t.Fatalf("Reallocation occured")
	}

	err = ipPool.Release(ip.IP)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	err = ipPool.LeaseWithAddr(ip)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
}

func TestLease(t *testing.T) {
	subnet, err := netlink.ParseAddr("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	ipPool := NewIP4AddrPool(subnet)

	ip1, err := ipPool.Lease()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	assert.Equal(t, ip1.String(), "192.168.1.1/24", "invalid addr")

	ip2, err := ipPool.Lease()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	assert.Equal(t, ip2.String(), "192.168.1.2/24", "invalid addr")

	ip3, err := ipPool.Lease()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	assert.Equal(t, ip3.String(), "192.168.1.3/24", "invalid addr")

	err = ipPool.Release(ip2.IP)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	ip4, err := ipPool.Lease()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	assert.Equal(t, ip4.String(), "192.168.1.2/24", "invalid addr")
}

func TestGetHostAddr(t *testing.T) {
	ip, err := netlink.ParseAddr("192.168.1.176/24")
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	hostAddr := int(getHostAddr(ip.IP, 24))
	assert.Equal(t, hostAddr, 176, "invalid hostAddr")
}
