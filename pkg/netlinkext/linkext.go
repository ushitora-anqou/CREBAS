package netlinkext

import (
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// LinkExt is an extension for netlink.Link
// +gen * slice:"Where"
type LinkExt struct {
	link         netlink.Link
	Addr         *netlink.Addr
	namespace    string
	OfType       OFType
	Ofport       uint32
	DefaultRoute bool
}

type OFType int

const (
	ACLOFSwitch OFType = iota
	ExternalOFSwitch
)

// NewLinkExtVeth creates veth LinkExt
func NewLinkExtVeth(linkName string, peerName string) *LinkExt {
	linkExt := &LinkExt{}

	vethLink := &netlink.Veth{}
	vethLink.LinkAttrs = netlink.NewLinkAttrs()
	vethLink.Name = linkName
	vethLink.PeerName = peerName
	linkExt.link = vethLink
	linkExt.DefaultRoute = false

	return linkExt
}

// Create link
func (l *LinkExt) Create() error {
	err := netlink.LinkAdd(l.link)
	if err != nil {
		return err
	}

	link, err := netlink.LinkByName(l.link.Attrs().Name)
	if err != nil {
		return err
	}
	vethLink := link.(*netlink.Veth)
	vethLink.PeerName = l.link.(*netlink.Veth).PeerName

	l.link = vethLink
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
		defer rootNetns.Close()

		err = netns.Set(handle)
		if err != nil {
			rootNetns.Close()
			return err
		}
		defer netns.Set(rootNetns)
	}

	err := netlink.AddrAdd(l.link, addr)
	if err != nil {
		return err
	}

	l.Addr = addr

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
		defer rootNetns.Close()

		err = netns.Set(handle)
		if err != nil {
			rootNetns.Close()
			return err
		}
		defer netns.Set(rootNetns)
	}

	err := netlink.LinkDel(l.link)
	return err
}

// GetLink returns link
func (l *LinkExt) GetLink() netlink.Link {
	return l.link
}

// SetLinkUp sets link up
func (l *LinkExt) SetLinkUp() error {
	if l.namespace != "" {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		rootNetns, err := netns.Get()
		if err != nil {
			return err
		}
		defer rootNetns.Close()

		handle, err := netns.GetFromName(l.namespace)
		if err != nil {
			return err
		}
		defer handle.Close()

		err = netns.Set(handle)
		if err != nil {
			rootNetns.Close()
			return err
		}
		defer netns.Set(rootNetns)
	}

	err := netlink.LinkSetUp(l.link)
	if err != nil {
		return err
	}

	return nil
}

// SetLinkPeerUp set veth peer link up
func (l *LinkExt) SetLinkPeerUp() error {
	peerName := l.link.(*netlink.Veth).PeerName

	link, err := netlink.LinkByName(peerName)
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}

	return nil
}

func (l *LinkExt) SetLink(link netlink.Link) {
	l.link = link
}
