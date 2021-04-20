package ofswitch

import (
	"fmt"
	"net"

	"github.com/naoki9911/gofc/ofprotocol/ofp13"
)

func (c *OFSwitch) AddHostAggregatedARPFlow(link DeviceLink) error {
	err := c.addHostAggregatedARPFlowHost(link)
	if err != nil {
		return err
	}
	err = c.addHostAggregatedARPFlowClientBroadcast(link)
	if err != nil {
		return err
	}
	err = c.addHostAggregatedARPFlowClientUnicast(link)
	if err != nil {
		return err
	}
	return nil
}

func (c *OFSwitch) DeleteHostAggregatedARPFlow(link DeviceLink) error {
	err := c.deleteHostAggregatedARPFlowHost(link)
	if err != nil {
		return err
	}
	err = c.deleteHostAggregatedARPFlowClientBroadcast(link)
	if err != nil {
		return err
	}
	err = c.deleteHostAggregatedARPFlowClientUnicast(link)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) getHostAggregatedARPFlowMatchingHost(link DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(c.Link.GetOfPort())
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(link.GetHWAddress().String())
	if err != nil {
		return nil, err
	}
	match.Append(ethsrc)

	ethType := ofp13.NewOxmEthType(0x806)
	if err != nil {
		return nil, err
	}
	match.Append(ethType)

	return match, nil
}

func (c *OFSwitch) addHostAggregatedARPFlowHost(link DeviceLink) error {
	match, err := c.getHostAggregatedARPFlowMatchingHost(link)
	if err != nil {
		return err
	}

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(link.GetOfPort(), OFPCML_NO_BUFFER))
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

func (c *OFSwitch) deleteHostAggregatedARPFlowHost(link DeviceLink) error {
	match, err := c.getHostAggregatedARPFlowMatchingHost(link)
	if err != nil {
		return err
	}

	fm := ofp13.NewOfpFlowModDelete(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		match,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) getHostAggregatedARPFlowMatchingClientBroadcast(link DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(link.GetOfPort())
	match.Append(inport)

	ethdst, err := ofp13.NewOxmEthDst("FF:FF:FF:FF:FF:FF")
	if err != nil {
		return nil, err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x806)
	if err != nil {
		return nil, err
	}
	match.Append(ethType)

	return match, nil
}
func (c *OFSwitch) addHostAggregatedARPFlowClientBroadcast(link DeviceLink) error {
	match, err := c.getHostAggregatedARPFlowMatchingClientBroadcast(link)
	if err != nil {
		return err
	}

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(c.Link.GetOfPort(), OFPCML_NO_BUFFER))
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

func (c *OFSwitch) deleteHostAggregatedARPFlowClientBroadcast(link DeviceLink) error {
	match, err := c.getHostAggregatedARPFlowMatchingClientBroadcast(link)
	if err != nil {
		return err
	}

	fm := ofp13.NewOfpFlowModDelete(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		match,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) getHostAggregatedARPFlowMatchingClientUnicast(link DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(link.GetOfPort())
	match.Append(inport)

	ethdst, err := ofp13.NewOxmEthDst(link.GetHWAddress().String())
	if err != nil {
		return nil, err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x806)
	if err != nil {
		return nil, err
	}
	match.Append(ethType)

	return match, nil
}
func (c *OFSwitch) addHostAggregatedARPFlowClientUnicast(link DeviceLink) error {
	match, err := c.getHostAggregatedARPFlowMatchingClientUnicast(link)
	if err != nil {
		return err
	}

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(c.Link.GetOfPort(), OFPCML_NO_BUFFER))
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

func (c *OFSwitch) deleteHostAggregatedARPFlowClientUnicast(link DeviceLink) error {
	match, err := c.getHostAggregatedARPFlowMatchingClientUnicast(link)
	if err != nil {
		return err
	}

	fm := ofp13.NewOfpFlowModDelete(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		match,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) AddHostAggregatedDHCPFlow(link DeviceLink) error {
	err := c.addHostAggregatedDHCPFlowHost(link)
	if err != nil {
		return err
	}
	err = c.addHostAggregatedDHCPFlowClientBroadcast(link)
	if err != nil {
		return err
	}
	err = c.addHostAggregatedDHCPFlowClientUnicast(link)
	if err != nil {
		return err
	}
	return nil
}

func (c *OFSwitch) DeleteHostAggregatedDHCPFlow(link DeviceLink) error {
	err := c.deleteHostAggregatedDHCPFlowHost(link)
	if err != nil {
		return err
	}
	err = c.deleteHostAggregatedDHCPFlowClientBroadcast(link)
	if err != nil {
		return err
	}
	err = c.deleteHostAggregatedDHCPFlowClientUnicast(link)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) getHostAggregatedDHCPFlowMatchingHost(link DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(c.Link.GetOfPort())
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(link.GetHWAddress().String())
	if err != nil {
		return nil, err
	}
	match.Append(ethsrc)

	ethType := ofp13.NewOxmEthType(0x0800)
	match.Append(ethType)

	fmt.Println(c.Link.GetIPAddress().IP.String())
	ipSrc, err := ofp13.NewOxmIpv4Src(c.Link.GetIPAddress().IP.String())
	if err != nil {
		return nil, err
	}
	match.Append(ipSrc)

	ipProto := ofp13.NewOxmIpProto(17)
	match.Append(ipProto)

	udpSrc := ofp13.NewOxmUdpSrc(67)
	match.Append(udpSrc)

	udpDst := ofp13.NewOxmUdpDst(68)
	match.Append(udpDst)

	return match, nil
}

func (c *OFSwitch) addHostAggregatedDHCPFlowHost(link DeviceLink) error {
	match, err := c.getHostAggregatedDHCPFlowMatchingHost(link)
	if err != nil {
		return err
	}

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(link.GetOfPort(), OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		200,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) deleteHostAggregatedDHCPFlowHost(link DeviceLink) error {
	match, err := c.getHostAggregatedDHCPFlowMatchingHost(link)
	if err != nil {
		return err
	}

	fm := ofp13.NewOfpFlowModDelete(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		match,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) getHostAggregatedDHCPFlowMatchingClientBroadcast(link DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(link.GetOfPort())
	match.Append(inport)

	ethdst, err := ofp13.NewOxmEthDst("FF:FF:FF:FF:FF:FF")
	if err != nil {
		return nil, err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x0800)
	match.Append(ethType)

	ipDst, err := ofp13.NewOxmIpv4Dst(net.IPv4bcast.String())
	if err != nil {
		return nil, err
	}
	match.Append(ipDst)

	ipProto := ofp13.NewOxmIpProto(17)
	match.Append(ipProto)

	udpSrc := ofp13.NewOxmUdpSrc(68)
	match.Append(udpSrc)

	udpDst := ofp13.NewOxmUdpDst(67)
	match.Append(udpDst)

	return match, nil
}

func (c *OFSwitch) addHostAggregatedDHCPFlowClientBroadcast(link DeviceLink) error {
	match, err := c.getHostAggregatedDHCPFlowMatchingClientBroadcast(link)
	if err != nil {
		return err
	}

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(c.Link.GetOfPort(), OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		200,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) deleteHostAggregatedDHCPFlowClientBroadcast(link DeviceLink) error {
	match, err := c.getHostAggregatedDHCPFlowMatchingClientBroadcast(link)
	if err != nil {
		return err
	}

	fm := ofp13.NewOfpFlowModDelete(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		match,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) getHostAggregatedDHCPFlowMatchingClientUnicast(link DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(link.GetOfPort())
	match.Append(inport)

	ethdst, err := ofp13.NewOxmEthDst(link.GetHWAddress().String())
	if err != nil {
		return nil, err
	}
	match.Append(ethdst)

	ethType := ofp13.NewOxmEthType(0x0800)
	match.Append(ethType)

	ipDst, err := ofp13.NewOxmIpv4Dst(c.Link.GetIPAddress().IP.String())
	if err != nil {
		return nil, err
	}
	match.Append(ipDst)

	ipProto := ofp13.NewOxmIpProto(17)
	match.Append(ipProto)

	udpSrc := ofp13.NewOxmUdpSrc(68)
	match.Append(udpSrc)

	udpDst := ofp13.NewOxmUdpDst(67)
	match.Append(udpDst)

	return match, nil
}

func (c *OFSwitch) addHostAggregatedDHCPFlowClientUnicast(link DeviceLink) error {
	match, err := c.getHostAggregatedDHCPFlowMatchingClientUnicast(link)
	if err != nil {
		return err
	}

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(c.Link.GetOfPort(), OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		200,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) deleteHostAggregatedDHCPFlowClientUnicast(link DeviceLink) error {
	match, err := c.getHostAggregatedDHCPFlowMatchingClientUnicast(link)
	if err != nil {
		return err
	}

	fm := ofp13.NewOfpFlowModDelete(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		match,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) SendFlowModAddOutput(match *ofp13.OfpMatch, outport uint32, priority uint16) error {
	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(outport, OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		priority,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) AddDeviceAppARPFlow(deviceLink DeviceLink, appLink DeviceLink) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(deviceLink.GetOfPort())
	match.Append(inport)

	ethType := ofp13.NewOxmEthType(0x0806)
	match.Append(ethType)

	arpSrc, err := ofp13.NewOxmArpSpa(deviceLink.GetIPAddress().IP.String())
	if err != nil {
		return err
	}
	match.Append(arpSrc)

	arpDst, err := ofp13.NewOxmArpTpa(appLink.GetIPAddress().IP.String())
	if err != nil {
		return err
	}
	match.Append(arpDst)

	err = c.SendFlowModAddOutput(match, appLink.GetOfPort(), 200)
	if err != nil {
		return err
	}

	match = ofp13.NewOfpMatch()

	inport = ofp13.NewOxmInPort(appLink.GetOfPort())
	match.Append(inport)

	ethType = ofp13.NewOxmEthType(0x0806)
	match.Append(ethType)

	arpSrc, err = ofp13.NewOxmArpSpa(appLink.GetIPAddress().IP.String())
	if err != nil {
		return err
	}
	match.Append(arpSrc)

	arpDst, err = ofp13.NewOxmArpTpa(deviceLink.GetIPAddress().IP.String())
	if err != nil {
		return err
	}
	match.Append(arpDst)

	err = c.SendFlowModAddOutput(match, deviceLink.GetOfPort(), 200)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) AddDeviceAppIPFlow(deviceLink DeviceLink, appLink DeviceLink) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(deviceLink.GetOfPort())
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(deviceLink.GetHWAddress().String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethType := ofp13.NewOxmEthType(0x0800)
	match.Append(ethType)

	ipSrc, err := ofp13.NewOxmIpv4Src(deviceLink.GetIPAddress().IP.String())
	if err != nil {
		return err
	}
	match.Append(ipSrc)

	err = c.SendFlowModAddOutput(match, appLink.GetOfPort(), 100)
	if err != nil {
		return err
	}

	match = ofp13.NewOfpMatch()

	inport = ofp13.NewOxmInPort(appLink.GetOfPort())
	match.Append(inport)

	ethsrc, err = ofp13.NewOxmEthSrc(appLink.GetHWAddress().String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	ethType = ofp13.NewOxmEthType(0x0800)
	match.Append(ethType)

	ipSrc, err = ofp13.NewOxmIpv4Src(appLink.GetIPAddress().IP.String())
	if err != nil {
		return err
	}
	match.Append(ipSrc)

	err = c.SendFlowModAddOutput(match, deviceLink.GetOfPort(), 100)
	if err != nil {
		return err
	}

	return nil
}

func (c *OFSwitch) AddDeviceAppTunnelFlow(deviceLink DeviceLink, appLink DeviceLink) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(deviceLink.GetOfPort())
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(deviceLink.GetHWAddress().String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(appLink.GetOfPort(), OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		10,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	match = ofp13.NewOfpMatch()

	inport = ofp13.NewOxmInPort(appLink.GetOfPort())
	match.Append(inport)

	ethdst, err := ofp13.NewOxmEthDst(deviceLink.GetHWAddress().String())
	if err != nil {
		return err
	}
	match.Append(ethdst)

	instruction = ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(deviceLink.GetOfPort(), OFPCML_NO_BUFFER))
	instructions = make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm = ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		10,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	match = ofp13.NewOfpMatch()

	inport = ofp13.NewOxmInPort(appLink.GetOfPort())
	match.Append(inport)

	ethdst, err = ofp13.NewOxmEthDst("FF:FF:FF:FF:FF:FF")
	if err != nil {
		return err
	}
	match.Append(ethdst)

	instruction = ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(deviceLink.GetOfPort(), OFPCML_NO_BUFFER))
	instructions = make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm = ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		10,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}
	return nil
}

func (c *OFSwitch) AddAppsTunnel(deviceLinkA DeviceLink, appLinkA DeviceLink, deviceLinkB DeviceLink, appLinkB DeviceLink) error {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(appLinkA.GetOfPort())
	match.Append(inport)

	ethsrc, err := ofp13.NewOxmEthSrc(deviceLinkA.GetHWAddress().String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(appLinkB.GetOfPort(), OFPCML_NO_BUFFER))
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm := ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		10,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	match = ofp13.NewOfpMatch()

	inport = ofp13.NewOxmInPort(appLinkB.GetOfPort())
	match.Append(inport)

	ethsrc, err = ofp13.NewOxmEthSrc(deviceLinkB.GetHWAddress().String())
	if err != nil {
		return err
	}
	match.Append(ethsrc)

	instruction = ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)
	instruction.Append(ofp13.NewOfpActionOutput(appLinkA.GetOfPort(), OFPCML_NO_BUFFER))
	instructions = make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	fm = ofp13.NewOfpFlowModAdd(
		0,
		0,
		0,
		10,
		0,
		match,
		instructions,
	)

	if !c.dp.Send(fm) {
		return fmt.Errorf("failed to send flow to switch(%v)", c.Name)
	}

	return nil
}

func (c *OFSwitch) getAppsARPMatch(deviceLinkA DeviceLink, appLinkA, deviceLinkB DeviceLink, appLinkB DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(appLinkA.GetOfPort())
	match.Append(inport)

	ethType := ofp13.NewOxmEthType(0x806)
	match.Append(ethType)

	arpSpa, err := ofp13.NewOxmArpSpa(deviceLinkA.GetIPAddress().IP.String())
	if err != nil {
		return nil, err
	}
	match.Append(arpSpa)

	arpTpa, err := ofp13.NewOxmArpTpa(deviceLinkB.GetIPAddress().IP.String())
	if err != nil {
		return nil, err
	}
	match.Append(arpTpa)

	return match, nil
}

func (c *OFSwitch) AddAppsARPFlow(deviceLinkA DeviceLink, appLinkA DeviceLink, deviceLinkB DeviceLink, appLinkB DeviceLink) error {
	matchA, err := c.getAppsARPMatch(deviceLinkA, appLinkA, deviceLinkB, appLinkB)
	if err != nil {
		return err
	}
	err = c.SendFlowModAddOutput(matchA, appLinkB.GetOfPort(), 90)
	if err != nil {
		return err
	}
	matchB, err := c.getAppsARPMatch(deviceLinkB, appLinkB, deviceLinkA, appLinkA)
	if err != nil {
		return err
	}
	err = c.SendFlowModAddOutput(matchB, appLinkA.GetOfPort(), 90)
	if err != nil {
		return err
	}
	return nil
}

func (c *OFSwitch) getAppsICMPMatch(deviceLinkA DeviceLink, appLinkA DeviceLink, deviceLinkB DeviceLink, appLinkB DeviceLink) (*ofp13.OfpMatch, error) {
	match := ofp13.NewOfpMatch()

	inport := ofp13.NewOxmInPort(appLinkA.GetOfPort())
	match.Append(inport)

	ethType := ofp13.NewOxmEthType(0x0800)
	match.Append(ethType)

	ipSrc, err := ofp13.NewOxmIpv4Src(deviceLinkA.GetIPAddress().IP.String())
	if err != nil {
		return nil, err
	}
	match.Append(ipSrc)

	ipDst, err := ofp13.NewOxmIpv4Dst(deviceLinkB.GetIPAddress().IP.String())
	if err != nil {
		return nil, err
	}
	match.Append(ipDst)

	ipProto := ofp13.NewOxmIpProto(1)
	match.Append(ipProto)

	return match, nil
}

func (c *OFSwitch) AddAppsICMPFlow(deviceLinkA DeviceLink, appLinkA DeviceLink, deviceLinkB DeviceLink, appLinkB DeviceLink) error {
	matchA, err := c.getAppsICMPMatch(deviceLinkA, appLinkA, deviceLinkB, appLinkB)
	if err != nil {
		return err
	}
	err = c.SendFlowModAddOutput(matchA, appLinkB.GetOfPort(), 90)
	if err != nil {
		return err
	}
	matchB, err := c.getAppsICMPMatch(deviceLinkB, appLinkB, deviceLinkA, appLinkA)
	if err != nil {
		return err
	}
	err = c.SendFlowModAddOutput(matchB, appLinkA.GetOfPort(), 90)
	if err != nil {
		return err
	}

	return nil
}
