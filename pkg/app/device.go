package app

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Device struct {
	HWAddress net.HardwareAddr `json:"hwAddress"`
	IPAddress *netlink.Addr    `json:"ipAddress"`
	App       AppInterface
}
