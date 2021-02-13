package netlinkext

import (
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// LinkExt is an extension for netlink.Link
// +gen * slice:"Where"
type LinkExt struct {
	link      netlink.Link
	addr      *netlink.Addr
	namespace string
}

// NewLinkExtVeth creates veth LinkExt
func NewLinkExtVeth(linkName string, peerName string) *LinkExt {
	linkExt := &LinkExt{}

	vethLink := &netlink.Veth{}
	vethLink.LinkAttrs = netlink.NewLinkAttrs()
	vethLink.Name = linkName
	vethLink.PeerName = peerName
	linkExt.link = vethLink

	return linkExt
}

// Create link
func (l *LinkExt) Create() error {
	err := netlink.LinkAdd(l.link)
	if err != nil {
		return err
	}

	return nil
}

// SetNsByName configure set into netns
func (l *LinkExt) SetNsByName(name string) error {
	handle, err := netns.GetFromName(name)
	if err != nil {
		return err
	}
	defer handle.Close()

	err = netlink.LinkSetNsFd(l.link, int(handle))
	if err != nil {
		return err
	}

	l.namespace = name

	return nil
}

// SetAddr configure link addr
func (l *LinkExt) SetAddr(addr *netlink.Addr) error {
	if l.namespace != "" {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		handle, err := netns.GetFromName(l.namespace)
		if err != nil {
			return err
		}
		defer handle.Close()

		rootNetns, err := netns.Get()
		if err != nil {
			return err
		}

		err = netns.Set(handle)
		if err != nil {
			rootNetns.Close()
			return err
		}
		defer netns.Set(rootNetns)
		defer rootNetns.Close()
	}

	err := netlink.AddrAdd(l.link, addr)
	if err != nil {
		return err
	}

	l.addr = addr
	return nil
}

// Delete link
func (l *LinkExt) Delete() error {
	if l.namespace != "" {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		handle, err := netns.GetFromName(l.namespace)
		if err != nil {
			return err
		}
		defer handle.Close()

		rootNetns, err := netns.Get()
		if err != nil {
			return err
		}

		err = netns.Set(handle)
		if err != nil {
			rootNetns.Close()
			return err
		}
		defer netns.Set(rootNetns)
		defer rootNetns.Close()
	}

	err := netlink.LinkDel(l.link)
	return err
}
