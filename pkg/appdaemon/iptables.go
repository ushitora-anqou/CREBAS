package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"

	"github.com/vishvananda/netlink"
)

func configureNAT(dgwLink netlink.Link, childLink netlink.Link) error {
	childAddrs, err := netlink.AddrList(childLink, netlink.FAMILY_V4)
	if err != nil {
		return err
	}

	if len(childAddrs) != 1 {
		return fmt.Errorf("invalid childLink %v", len(childAddrs))
	}

	childAddr := childAddrs[0]
	childAddrSubnet := childAddr.IPNet
	childLinkName := childLink.Attrs().Name
	dgwLinkName := dgwLink.Attrs().Name

	if !existMasqueradeRule(childAddrSubnet, dgwLinkName) {
		log.Printf("info: Adding MASQUERADE Rule")
		err := addMasqueradeRule(childAddrSubnet, dgwLinkName)
		if err != nil {
			panic(err)
		}
	}

	if !existForwardRule(childLinkName, childAddrSubnet, dgwLinkName, "") {
		log.Printf("info: Adding FORWARD Rule")
		err := addForwardRule(childLinkName, childAddrSubnet, dgwLinkName, "")
		if err != nil {
			panic(err)
		}
	}

	if existForwardRule("", nil, "", "RELATED,ESTABLISHED") {
		log.Printf("info: Adding Stateful FORWARD Rule")
		err := addForwardRule("", nil, "", "RELATED,ESTABLISHED")
		if err != nil {
			panic(err)
		}
	}

	log.Printf("info: Successfully configured NAT")

	return nil
}

func addMasqueradeRule(inboundNetwork *net.IPNet, outboundDevice string) error {
	subnetLength, _ := inboundNetwork.Mask.Size()
	networkIP := inboundNetwork.IP.String() + "/" + strconv.Itoa(subnetLength)
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", networkIP, "-o", outboundDevice, "-j", "MASQUERADE")
	return cmd.Run()
}

func existMasqueradeRule(inboundNetwork *net.IPNet, outboundDevice string) bool {
	subnetLength, _ := inboundNetwork.Mask.Size()
	networkIP := inboundNetwork.IP.String() + "/" + strconv.Itoa(subnetLength)
	fmt.Println(networkIP)
	cmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", networkIP, "-o", outboundDevice, "-j", "MASQUERADE")
	cmd.Run()

	if cmd.ProcessState.ExitCode() != 0 {
		return false
	}
	return true
}

func addForwardRule(inboundDevice string, inboundNetwork *net.IPNet, outboundDevice string, state string) error {
	var cmd *exec.Cmd
	if state == "" {
		subnetLength, _ := inboundNetwork.Mask.Size()
		networkIP := inboundNetwork.IP.String() + "/" + strconv.Itoa(subnetLength)
		cmd = exec.Command("iptables", "-A", "FORWARD", "-i", inboundDevice, "-s", networkIP, "-o", outboundDevice, "-j", "ACCEPT")
	} else if inboundDevice == "" && inboundNetwork == nil && outboundDevice == "" {
		cmd = exec.Command("iptables", "-A", "FORWARD", "-m", "conntrack", "--ctstate", state, "-j", "ACCEPT")
	} else {
		return fmt.Errorf("undefined command")
	}
	return cmd.Run()
}

func existForwardRule(inboundDevice string, inboundNetwork *net.IPNet, outboundDevice string, state string) bool {
	var cmd *exec.Cmd
	if state == "" {
		subnetLength, _ := inboundNetwork.Mask.Size()
		networkIP := inboundNetwork.IP.String() + "/" + strconv.Itoa(subnetLength)
		cmd = exec.Command("iptables", "-C", "FORWARD", "-i", inboundDevice, "-s", networkIP, "-o", outboundDevice, "-j", "ACCEPT")
	} else if inboundDevice == "" && inboundNetwork == nil && outboundDevice == "" {
		cmd = exec.Command("iptables", "-C", "FORWARD", "-m", "conntrack", "--ctstate", state, "-j", "ACCEPT")
	} else {
		log.Printf("error: Undefined command")
		return false
	}

	if cmd.Run(); cmd.ProcessState.ExitCode() != 0 {
		return false
	}
	return true
}
