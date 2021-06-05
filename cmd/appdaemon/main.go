package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/vishvananda/netlink"
)

var childCmd *exec.Cmd
var device *app.Device
var appInfo *app.AppInfo
var grantedCapabilities *capability.CapabilityCollection = capability.NewCapabilityCollection()
var ovsInfo *ofswitch.OvsInfo
var certificate *x509.Certificate
var privateKey *rsa.PrivateKey

var cpCert *capability.AppCertificate
var userCert *capability.AppCertificate

var tempAllowed = false
var humidAllowed = false

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) > 1 && args[0] == "testMode" {
		testMode(args[1])
	}

	pkgInfo, err := pkg.OpenPackageInfo(args[0])
	if err != nil {
		panic(err)
	}

	fmt.Printf("CertificatePath: %v\n", pkgInfo.CertificatePath)
	fmt.Printf("PrivateKeyPath : %v\n", pkgInfo.PrivateKeyPath)
	certBytes, err := capability.ReadCertificateWithoutDecode(pkgInfo.CertificatePath)
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	certificate, err = capability.DecodeCertificate(certBytes)
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	privateKey, err = capability.ReadPrivateKey(pkgInfo.PrivateKeyPath)
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	output, _ := exec.Command("/bin/ip", "a", "s").Output()
	fmt.Println(string(output))

	defaultRoute, err := getDefaultRoute()
	if err != nil {
		fmt.Println("Default route not found")
	} else {
		fmt.Printf("Default route: %v\n", defaultRoute.String())
		appID, err := uuid.Parse(args[1])
		if err != nil {
			fmt.Println(err)
		} else {
			pepUrl := "http://" + defaultRoute.String() + ":8080"
			appInfo, err = getAppInfo(appID, pepUrl)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(appInfo)
			}

			cpUrl := "http://" + defaultRoute.String() + ":8081"
			for idx := range pkgInfo.Capabilities {
				pkgInfo.Capabilities[idx].AppID = appID
				pkgInfo.Capabilities[idx].AssignerID = appID
				err = pkgInfo.Capabilities[idx].Sign(privateKey)
				if err != nil {
					fmt.Println(err)
				}
			}
			for idx := range pkgInfo.CapabilityRequests {
				pkgInfo.CapabilityRequests[idx].RequesterID = appID
				err = pkgInfo.CapabilityRequests[idx].Sign(privateKey)
				if err != nil {
					fmt.Println(err)
				}
			}

			device, err = getAppDevice(appID, pepUrl)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(device)
			}

			ovsInfo, err = getOvsInfo(pepUrl)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(device)
			}

			certBase64 := base64.StdEncoding.EncodeToString(certBytes)
			appCert := capability.AppCertificate{
				AppID:             appID,
				CertificateString: certBase64,
			}
			_, err = capability.SendContentsToCP(cpUrl+"/app/cert", appCert)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			cpCert, err = getCertificate(cpUrl + "/app/cpCert")
			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			userCert, err = getCertificate(cpUrl + "/app/userCert")
			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			fmt.Printf("DeviceLinkName:%v ACLLinkName:%v\n", appInfo.DeviceLinkName, appInfo.ACLLinkName)
			go func() {
				for {
					fmt.Println("Processing capabilities")
					grantedCaps, err := procCapability(appID, pkgInfo, cpUrl)
					if err != nil {
						fmt.Printf("error: failed to proc cap %v\n", err)
					}

					for idx := range grantedCaps {
						grantedCap := grantedCaps[idx]
						if grantedCap.AssignerID == cpCert.AppID {
							if grantedCap.Verify(cpCert.Certificate.PublicKey.(*rsa.PublicKey)) != nil {
								fmt.Printf("error: Failed to verify %v with cp cert\n", grantedCap.CapabilityID)
								continue
							}
						} else if grantedCap.AssignerID == userCert.AppID {
							if grantedCap.Verify(userCert.Certificate.PublicKey.(*rsa.PublicKey)) != nil {
								fmt.Printf("error: Failed to verify %v with user cert\n", grantedCap.CapabilityID)
								continue
							}
						} else {
							fmt.Printf("Unexpected AssignerID %v\n", grantedCap.AssignerID)
							continue
						}
						if grantedCapabilities.Contains(grantedCap) {
							continue
						}
						grantedCapabilities.Add(grantedCap)
						fmt.Printf("Enforce Cap %v\n", grantedCap.CapabilityID)
						_, err = capability.SendContentsToCP(pepUrl+"/app/"+appID.String()+"/cap", grantedCap)
						if err != nil {
							fmt.Printf("error: failed to send granted cap %v\n", err)
						}

						if grantedCap.CapabilityName == capability.CAPABILITY_NAME_TEMPERATURE {
							tempAllowed = true
						}

						if grantedCap.CapabilityName == capability.CAPABILITY_NAME_HUMIDITY {
							humidAllowed = true
						}

					}

					time.Sleep(1 * time.Second)
				}
			}()

			if !pkgInfo.TestUse {
				go startPassing(appInfo.DeviceLinkName, appInfo.ACLLinkName, true)
			}

		}
	}

	fmt.Println("Starting appdaemon...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go signalHandler(signalChan)

	start(pkgInfo)
	exitCode := wait()

	output, _ = exec.Command("/bin/ip", "a", "s").Output()
	fmt.Println(string(output))

	fmt.Printf("Exiting appdaemon %v\n", exitCode)
	os.Exit(exitCode)
}

func procCapability(appID uuid.UUID, pkgInfo *pkg.PackageInfo, cpUrl string) (capability.CapabilitySlice, error) {
	_, err := capability.SendContentsToCP(cpUrl+"/cap", pkgInfo.Capabilities)
	if err != nil {
		return nil, err
	}
	grantedCaps := capability.CapabilitySlice{}
	for idx := range pkgInfo.CapabilityRequests {
		capsByte, err := capability.SendContentsToCP(cpUrl+"/capReq", pkgInfo.CapabilityRequests[idx])
		if err != nil {
			return nil, err
		}
		var grantedCap capability.CapReqResponse
		err = json.Unmarshal(capsByte, &grantedCap)
		if err != nil {
			return nil, err
		}

		for idx := range grantedCap.GrantedCapabilities {
			grantedCaps = append(grantedCaps, grantedCap.GrantedCapabilities[idx])
		}
	}

	return grantedCaps, nil
}

func start(pkgInfo *pkg.PackageInfo) error {

	buff := make([]byte, 1024)

	cmds := pkgInfo.MetaInfo.CMD
	childCmd = exec.Command(cmds[0], cmds[1:]...)
	stdout, _ := childCmd.StdoutPipe()

	err := childCmd.Start()
	if err != nil {
		return err
	}

	go func() {
		n, err := stdout.Read(buff)

		for err == nil || err != io.EOF {
			if n > 0 {
				fmt.Print(string(buff[:n]))
			}

			n, err = stdout.Read(buff)
		}
	}()

	return nil
}

func stop() int {
	syscall.Kill(childCmd.Process.Pid, syscall.SIGTERM)
	return wait()
}

func wait() int {
	childCmd.Wait()
	return childCmd.ProcessState.ExitCode()
}

func signalHandler(signalChan chan os.Signal) {
	s := <-signalChan
	switch s {
	case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		syscall.Kill(childCmd.Process.Pid, syscall.SIGTERM)
	}
}

func testMode(mode string) {
	exec.Command("rm", "/tmp/appdaemon_test").Run()
	pkgInfo := pkg.CreateSkeltonPackageInfo()
	if mode == "noSIGTERM" {
		pkgInfo.MetaInfo.CMD = []string{"echo", "hello"}
	} else if mode == "SIGTERM" {
		pkgInfo.MetaInfo.CMD = []string{"sleep", "10"}
	} else {
		fmt.Printf("invalid argument %v\n", mode)
	}

	start(pkgInfo)
	exitCode := wait()
	file, err := os.Create("/tmp/appdaemon_test")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	file.WriteString(strconv.Itoa(exitCode) + "\n")
	file.Close()
	os.Exit(exitCode)
}

func getAppInfo(appID uuid.UUID, url string) (*app.AppInfo, error) {
	urlApp := url + "/app/" + appID.String()
	fmt.Println(urlApp)
	resp, err := http.Get(urlApp)
	if err != nil {
		return nil, err
	}
	fmt.Println("HOGE" + urlApp)

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	appInfo := app.AppInfo{}
	err = json.Unmarshal(respByte, &appInfo)
	if err != nil {
		return nil, err
	}

	if appID != appInfo.Id {
		return nil, fmt.Errorf("invalid appID")
	}

	return &appInfo, nil
}

func getAppDevice(appID uuid.UUID, url string) (*app.Device, error) {
	urlApp := url + "/app/" + appID.String() + "/device"
	fmt.Println(urlApp)
	resp, err := http.Get(urlApp)
	if err != nil {
		return nil, err
	}
	fmt.Println("HOGE" + urlApp)

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	device := app.Device{}
	err = json.Unmarshal(respByte, &device)
	if err != nil {
		return nil, err
	}

	return &device, nil
}

func getCertificate(url string) (*capability.AppCertificate, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	appCert := capability.AppCertificate{}
	err = json.Unmarshal(respByte, &appCert)
	if err != nil {
		return nil, err
	}

	err = appCert.Decode()
	if err != nil {
		return nil, err
	}

	return &appCert, nil
}

func getOvsInfo(url string) (*ofswitch.OvsInfo, error) {
	urlOvs := url + "/ovs"
	fmt.Println(urlOvs)
	resp, err := http.Get(urlOvs)
	if err != nil {
		return nil, err
	}
	fmt.Println("HOGE" + urlOvs)

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ovsInfo := ofswitch.OvsInfo{}
	err = json.Unmarshal(respByte, &ovsInfo)
	if err != nil {
		return nil, err
	}

	return &ovsInfo, nil

}

func getDefaultRoute() (net.IP, error) {
	routeList, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	for idx := range routeList {
		route := routeList[idx]
		if route.Scope == netlink.SCOPE_UNIVERSE {
			return route.Gw, nil
		}
	}

	return nil, fmt.Errorf("default route not found")
}

type Data struct {
	Opcode int     `json:"opcode"`
	Value  float64 `json:"value"`
}

func startPassing(recvLinkName string, sendLinkName string, recvIsDevice bool) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 0x0300)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer syscall.Close(fd)
	err = syscall.BindToDevice(fd, recvLinkName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	recvLink, err := netlink.LinkByName(recvLinkName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	sendLink, err := netlink.LinkByName(sendLinkName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	data := make([]byte, 1600)
	fmt.Printf("Device: %v SendVeth: %v RecvVeth:%v\n", device.HWAddress, appInfo.ACLLinkPeerHWAddress, appInfo.DeviceLinkPeerHWAddress)
	fmt.Printf("OvsACL: %v OvsExt: %v\n", ovsInfo.OvsACLHWAddr, ovsInfo.OvsExtHWAddr)
	for {
		n, addr, err := syscall.Recvfrom(fd, data, 0)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}
		packet := gopacket.NewPacket(data[0:n], layers.LayerTypeEthernet, gopacket.Default)
		ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		wifiMACStr := "f4:8c:50:30:da:4a"
		wifiMAC, _ := net.ParseMAC(wifiMACStr)
		if ethernetPacket.DstMAC.String() == wifiMAC.String() || ethernetPacket.SrcMAC.String() == wifiMAC.String() {
			continue
		}
		if ethernetPacket.DstMAC.String() == recvLink.Attrs().HardwareAddr.String() || ethernetPacket.SrcMAC.String() == recvLink.Attrs().HardwareAddr.String() {
			continue
		}
		if ethernetPacket.DstMAC.String() == sendLink.Attrs().HardwareAddr.String() || ethernetPacket.SrcMAC.String() == sendLink.Attrs().HardwareAddr.String() {
			continue
		}
		if ethernetPacket.DstMAC.String() == appInfo.ACLLinkPeerHWAddress || ethernetPacket.SrcMAC.String() == appInfo.ACLLinkPeerHWAddress {
			continue
		}
		if ethernetPacket.DstMAC.String() == appInfo.DeviceLinkPeerHWAddress || ethernetPacket.SrcMAC.String() == appInfo.DeviceLinkPeerHWAddress {
			continue
		}
		if ethernetPacket.DstMAC.String() == ovsInfo.OvsACLHWAddr || ethernetPacket.SrcMAC.String() == ovsInfo.OvsACLHWAddr {
			continue
		}
		if ethernetPacket.DstMAC.String() == ovsInfo.OvsExtHWAddr || ethernetPacket.SrcMAC.String() == ovsInfo.OvsExtHWAddr {
			continue
		}
		ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
		if ipv6Layer != nil {
			continue
		}
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udpPacket, _ := udpLayer.(*layers.UDP)
			if udpPacket.DstPort == 8000 {
				fmt.Printf("src port %d to dst to %d\n", udpPacket.SrcPort, udpPacket.DstPort)
				applicationLayer := packet.ApplicationLayer()
				fmt.Printf("%s\n", applicationLayer.Payload())
				var data = Data{}
				json.Unmarshal(applicationLayer.Payload(), &data)
				fmt.Println(data)
				if !appInfo.Server {
					if data.Opcode == 0 {
						if !tempAllowed {
							fmt.Println("temp not allowed")
							continue
						}
					} else if data.Opcode == 1 {
						if !humidAllowed {
							fmt.Println("humid not allowed")
							continue
						}
					}
				}

			}
		}

		//fmt.Printf("Recv IF NAME: %v Send IF NAME: %v", recvLinkName, sendLinkName)
		//fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
		//fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
		err = syscall.Sendto(fd, data[0:n], 0, addr)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}

		//fmt.Println(data)
	}
}
