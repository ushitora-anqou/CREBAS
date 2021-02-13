package netlinkext

import (
	"github.com/vishvananda/netlink"
)

// CreateVethPeer creates veth peer
func CreateVethPeer(linkName string, peerName string) error {
	vethLink := &netlink.Veth{}
	vethLink.LinkAttrs = netlink.NewLinkAttrs()
	vethLink.Name = linkName
	vethLink.PeerName = peerName

	return netlink.LinkAdd(vethLink)
}
