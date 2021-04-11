package ofswitch

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/digitalocean/go-openvswitch/ovs"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/gofc"
	"github.com/naoki9911/gofc/ofprotocol/ofp13"
	"github.com/vishvananda/netlink"
)

// OFSwitch is a struct to manage openflow switch
type OFSwitch struct {
	Name          string
	ControllerURL string
	client        *ovs.Client
	link          *netlinkext.LinkExt
	ports         *netlinkext.LinkCollection
	datapathID    uint64
	dp            *gofc.Datapath
}

// NewOFSwitch creates openflow switch
func NewOFSwitch(switchName string) *OFSwitch {
	ofs := new(OFSwitch)

	ofs.Name = switchName
	ofs.client = ovs.New()
	ofs.ports = netlinkext.NewLinkCollection()
	ofs.datapathID = 0
	ofs.dp = nil
	ofs.link = &netlinkext.LinkExt{
		Ofport: ofPortLocal,
	}

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
	s.link.SetLink(link)

	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}

	// Get DatapathID to controll the bridge with Ryu
	out, err := exec.Command("ovs-vsctl", "get", "bridge", s.Name, "datapath-id").Output()
	if err != nil {
		log.Printf("error: Failed to get %v DatapthID", s.Name)
		return err
	}

	err = exec.Command("ovs-vsctl", "set", "bridge", s.Name, "protocols=OpenFlow13").Run()
	if err != nil {
		log.Printf("error: Failed to set %v version OpenFlow 1.3", s.Name)
		return err
	}

	// format '"xxxxxx(datapathID)"'
	datapathIDStr := strings.Trim(string(out), "\n")
	s.datapathID, err = strconv.ParseUint(datapathIDStr[1:len(datapathIDStr)-1], 16, 64)
	if err != nil {
		return err
	}

	return nil
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
	err := netlink.AddrAdd(s.link.GetLink(), addr)
	if err != nil {
		return err
	}
	s.link.Addr = addr
	return nil
}

// HandleSwitchFeatures handle ovs features
func (c *OFSwitch) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, dp *gofc.Datapath) {
	if msg.DatapathId != c.datapathID {
		return
	}
	c.dp = dp
	fmt.Println("Handle SwitchFeatures")
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

func (c *OFSwitch) HandleErrorMsg(msg *ofp13.OfpErrorMsg, dp *gofc.Datapath) {
	log.Printf("error: HandleErrorMsg Type:%d Code:%d", msg.Type, msg.Code)
}

func (c *OFSwitch) HandlePortStatus(msg *ofp13.OfpPortStatus, dp *gofc.Datapath) {
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
		return fmt.Errorf("unknown link type:%T", link)
	}

	c.ports.Add(linkExt)
	return nil
}

func (c *OFSwitch) IsConnectedToController() bool {
	return c.dp != nil
}

func (c *OFSwitch) AddHostRestrictedFlow(link *netlinkext.LinkExt) error {
	err := c.AddARPFlow(link, c.link)
	if err != nil {
		return err
	}

	err = c.AddICMPFlow(link, c.link)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) AddARPFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	err := c.addUnicastARPFlow(linkA, linkB)
	if err != nil {
		return err
	}

	err = c.addUnicastARPFlow(linkB, linkA)
	if err != nil {
		return err
	}

	err = c.addBroadcastARPFlow(linkA, linkB)
	if err != nil {
		return err
	}

	err = c.addBroadcastARPFlow(linkB, linkA)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) addUnicastARPFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(linkA.Ofport)
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(linkA.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethdst, err := ofp13.NewOxmEthDst(linkB.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x806)
	if err != nil {
		return err
	}
	match.Append(ethType)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(linkB.Ofport, OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		0,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) addBroadcastARPFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(linkA.Ofport)
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(linkA.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethdst, err := ofp13.NewOxmEthDst("FF:FF:FF:FF:FF:FF")
	if err != nil {
		return err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x806)
	if err != nil {
		return err
	}
	match.Append(ethType)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(linkB.Ofport, OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		0,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) AddICMPFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	err := c.addUnicastICMPFlow(linkA, linkB)
	if err != nil {
		return err
	}

	err = c.addUnicastICMPFlow(linkB, linkA)
	if err != nil {
		return err
	}

	return nil
}

const ofPortLocal = 0xfffffffe
const OFPCML_NO_BUFFER = 0xffff

func (c *OFSwitch) addUnicastICMPFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(linkA.Ofport)
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(linkA.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethdst, err := ofp13.NewOxmEthDst(linkB.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x800)
	if err != nil {
		return err
	}
	match.Append(ethType)

	ipSrc, err := ofp13.NewOxmIpv4Src(linkA.Addr.IP.String())
	if err != nil {
		return err
	}
	match.Append(ipSrc)

	ipDst, err := ofp13.NewOxmIpv4Dst(linkB.Addr.IP.String())
	if err != nil {
		return err
	}
	match.Append(ipDst)

	ipProto := ofp13.NewOxmIpProto(1)
	match.Append(ipProto)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(linkB.Ofport, OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		0,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) AddTunnelFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	err := c.addUnicastTunnelFlow(linkA, linkB)
	if err != nil {
		return err
	}

	err = c.addUnicastTunnelFlow(linkB, linkA)
	if err != nil {
		return err
	}

	err = c.addBroadcastTunnelFlow(linkA, linkB)
	if err != nil {
		return err
	}

	err = c.addBroadcastTunnelFlow(linkB, linkA)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) addUnicastTunnelFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(linkA.Ofport)
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(linkA.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethdst, err := ofp13.NewOxmEthDst(linkB.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethdst)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(linkB.Ofport, OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		0,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) addBroadcastTunnelFlow(linkA *netlinkext.LinkExt, linkB *netlinkext.LinkExt) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(linkA.Ofport)
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(linkA.GetLink().Attrs().HardwareAddr.String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethdst, err := ofp13.NewOxmEthDst("FF:FF:FF:FF:FF:FF")
	if err != nil {
		return err
	}
	match.Append(ethdst)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(linkB.Ofport, OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		0,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}
