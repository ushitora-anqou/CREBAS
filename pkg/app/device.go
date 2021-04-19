package app

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Device struct {
	HWAddress net.HardwareAddr `json:"hwAddress"`
	IPAddress *netlink.Addr    `json:"ipAddress"`
	App       AppInterface
	OfPort    uint32
	ViaWlan   bool
}

func (d *Device) GetHWAddress() net.HardwareAddr {
	return d.HWAddress
}

func (d *Device) GetIPAddress() *netlink.Addr {
	return d.IPAddress
}

func (d *Device) GetOfPort() uint32 {
	return d.OfPort
}

func (d *Device) GetViaWlan() bool {
	return d.ViaWlan
}
