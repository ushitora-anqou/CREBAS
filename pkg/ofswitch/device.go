package ofswitch

import (
	"net"

	"github.com/vishvananda/netlink"
)

type DeviceLink interface {
	GetHWAddress() net.HardwareAddr
	GetIPAddress() *netlink.Addr
	GetOfPort() uint32
	GetViaWlan() bool
}
