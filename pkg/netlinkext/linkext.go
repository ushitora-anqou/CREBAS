package netlinkext

import (
	"github.com/vishvananda/netlink"
)

// LinkExt is an extension for netlink.Link
// +gen * slice:"Where"
type LinkExt struct {
	link netlink.Link
}
