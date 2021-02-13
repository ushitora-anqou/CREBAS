package netlinkext

import (
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func CreateNetns(name string) (netns.NsHandle, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rootNetns, err := netns.Get()
	if err != nil {
		return 0, err
	}
	defer rootNetns.Close()

	handle, err := netns.NewNamed(name)
	if err != nil {
		return 0, err
	}

	err = netns.Set(rootNetns)
	if err != nil {
		return 0, err
	}

	return handle, nil
}

// GetLinkByNameWithNetns returns link by name in specified netns
func GetLinkByNameWithNetns(name string, handle netns.NsHandle) (netlink.Link, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rootNetns, err := netns.Get()
	if err != nil {
		return nil, err
	}
	defer rootNetns.Close()

	err = netns.Set(handle)
	if err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	err = netns.Set(rootNetns)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func GetLinkAddrWithNetns(link netlink.Link, family int, handle netns.NsHandle) ([]netlink.Addr, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rootNetns, err := netns.Get()
	if err != nil {
		return nil, err
	}
	defer rootNetns.Close()

	err = netns.Set(handle)
	if err != nil {
		return nil, err
	}

	addrs, err := netlink.AddrList(link, family)
	if err != nil {
		return nil, err
	}

	err = netns.Set(rootNetns)
	if err != nil {
		return nil, err
	}

	return addrs, nil
}
