package ofswitch

import (
	"fmt"

	"github.com/Kmotiko/gofc"
	"github.com/Kmotiko/gofc/ofprotocol/ofp13"
	"github.com/digitalocean/go-openvswitch/ovs"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/vishvananda/netlink"
)

// OFSwitch is a struct to manage openflow switch
type OFSwitch struct {
	Name          string
	ControllerURL string
	IPAddr        *netlink.Addr
	client        *ovs.Client
	link          netlink.Link
	ports         *netlinkext.LinkCollection
}

// NewOFSwitch creates openflow switch
func NewOFSwitch(switchName string) *OFSwitch {
	ofs := new(OFSwitch)

	ofs.Name = switchName
	ofs.client = ovs.New()
	ofs.ports = netlinkext.NewLinkCollection()

	return ofs
}

// Create ovs
func (s *OFSwitch) Create() error {
	err := s.client.VSwitch.AddBridge(s.Name)
	if err != nil {
		return err
	}

	link, err := netlink.LinkByName(s.Name)
	if err != nil {
		return err
	}
	s.link = link

	err = netlink.LinkSetUp(link)
	return err
}

// Delete ovs
func (s *OFSwitch) Delete() error {
	return s.client.VSwitch.DeleteBridge(s.Name)
}

// SetController for ovs
func (s *OFSwitch) SetController(controllerURL string) error {
	s.ControllerURL = controllerURL
	return s.client.VSwitch.SetController(s.Name, s.ControllerURL)
}

// SetAddr configure ip(v4/v6) for ovs
func (s *OFSwitch) SetAddr(addr *netlink.Addr) error {
	err := netlink.AddrAdd(s.link, addr)
	if err != nil {
		return err
	}
	s.IPAddr = addr
	return nil
}

// HandleSwitchFeatures handle ovs features
func (c *OFSwitch) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, dp *gofc.Datapath) {
	// create match
	ethdst, _ := ofp13.NewOxmEthDst("00:00:00:00:00:00")
	if ethdst == nil {
		fmt.Println(ethdst)
		return
	}
	match := ofp13.NewOfpMatch()
	match.Append(ethdst)

	// create Instruction
	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)

	// create actions
	seteth, _ := ofp13.NewOxmEthDst("11:22:33:44:55:66")
	instruction.Append(ofp13.NewOfpActionSetField(seteth))

	// append Instruction
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	// create flow mod
	fm := ofp13.NewOfpFlowModModify(
		0, // cookie
		0, // cookie mask
		0, // tableid
		0, // priority
		ofp13.OFPFF_SEND_FLOW_REM,
		match,
		instructions,
	)

	// send FlowMod
	dp.Send(fm)

	// Create and send AggregateStatsRequest
	mf := ofp13.NewOfpMatch()
	mf.Append(ethdst)
	mp := ofp13.NewOfpAggregateStatsRequest(0, 0, ofp13.OFPP_ANY, ofp13.OFPG_ANY, 0, 0, mf)
	dp.Send(mp)
}

// HandleAggregateStatsReply reply some
func (c *OFSwitch) HandleAggregateStatsReply(msg *ofp13.OfpMultipartReply, dp *gofc.Datapath) {
	fmt.Println("Handle AggregateStats")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpAggregateStats); ok {
			fmt.Println(obj.PacketCount)
			fmt.Println(obj.ByteCount)
			fmt.Println(obj.FlowCount)
		}
	}
}

// AttackLink attaches link to ovs
func (c *OFSwitch) AttachLink(linkExt *netlinkext.LinkExt) error {
	switch link := linkExt.GetLink().(type) {
	case *netlink.Veth:
		c.client.VSwitch.AddPort(c.Name, link.PeerName)
		ofport, err := getOFPortByLinkName(link.PeerName)
		if err != nil {
			return err
		}
		linkExt.Ofport = ofport
	default:
		return fmt.Errorf("Unknown link type:%T", link)
	}

	c.ports.Add(linkExt)
	return nil
}
